package main

import (
	"errors"
	"log"
	"sync/atomic"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/types"
)

type Readable interface {
	Stream
	Read() types.Chunk
}

type ReadableStream struct {
	// embedded event emitter
	emitter.Emitter
	gen  generators.Type
	Type StreamType
	// internal read method that can be overwritten
	read  func()
	state *InternalState
}

func (rs *ReadableStream) getInternalState() *InternalState {
	return rs.state
}

func acquireRead(r Readable) bool {
	state := r.getInternalState()
	return atomic.CompareAndSwapUint32(&state.reading, 0, 1)
}

func releaseRead(r Readable) bool {
	state := r.getInternalState()
	return atomic.CompareAndSwapUint32(&state.reading, 1, 0)
}

func isReading(r Readable) bool {
	state := r.getInternalState()
	return atomic.LoadUint32(&state.reading) == 1
}

func NewReadable(conf StreamConfig) (r Readable, err error) {
	rs := &ReadableStream{
		read:  conf.read,
		gen:   conf.generator,
		state: NewInternalState(conf),
	}

	gen := conf.generator

	if gen == nil {
		return r, errors.New("readable stream generator is requred")
	}

	rs.read = func() {
		if !acquireRead(rs) {
			return
		}

		go func() {
			for canReadMore(rs) {
				// Get the next chunk
				chunk, err := gen.Next()

				if err != nil {
					log.Fatal(err.Error())
				}

				if chunk != nil {
					rs.state.buffer <- chunk
				} else {
					readableEndOfChunk(rs)
				}
			}

			if !releaseRead(rs) {
				log.Fatal("Could not unset reading flag")
			}
		}()
	}

	rs.On("data", func(evt types.Event) {
		if !rs.state.dest.Write(evt.Data) {
			pauseReadable(rs)
			if atomic.CompareAndSwapUint32(&rs.state.awaitDrainWriters, 0, 1) {
				rs.state.dest.Once("drain", func(evt types.Event) {
					atomic.CompareAndSwapUint32(&rs.state.awaitDrainWriters, 1, 0)
					resumeReadable(rs)
				})
			}
		}
	})

	rs.Once("readable_end", func(evt types.Event) {
		if readableDestroyed(rs) {
			return
		}
		state := rs.state
		// Write a nil chunk to the destination to signal end
		state.dest.Write(nil)

		if !atomic.CompareAndSwapInt32(&state.readableDestroyed, 0, 1) {
			log.Fatal("could not destroy readable - already destroyed")
		}
		// Call close on doneChan to send signal to the channel
		close(state.doneCh)
	})

	return rs, err
}

func readableDestroyed(r Readable) bool {
	state := r.getInternalState()
	return atomic.LoadInt32(&state.readableDestroyed) == 1
}

func readableEnded(r Readable) bool {
	state := r.getInternalState()
	return atomic.LoadInt32(&state.readableEnded) == 1
}

func canReadMore(r Readable) bool {
	if readableEnded(r) {
		return false
	}

	return true
}

// Set the ended flag
func readableEndOfChunk(r Readable) {
	if readableEnded(r) {
		log.Fatal("Cannot call readableEndOfChunk(), readable already ended")
	}

	if !endReadable(r) {
		log.Fatal("Could not end readable - already ended")
	}
}

func endReadable(r Readable) bool {
	state := r.getInternalState()
	if atomic.CompareAndSwapInt32(&state.readableEnded, 0, 1) {
		close(state.buffer)
		return true
	}
	return false
}

func readableFlowing(r Readable) bool {
	state := r.getInternalState()
	return atomic.LoadUint32(&state.mode) == ReadableFlowing
}

func emitData(r Readable, data types.Chunk) {
	r.Emit("data", data)
}

func (rs *ReadableStream) Done() <-chan struct{} {
	return rs.state.doneCh
}

// Read is the main method to call in order to either read some
// data from the generator and / or emit a readable event to
// write some data to the destination.
func (rs *ReadableStream) Read() (data types.Chunk) {
	// if stream is destroyed then there is no point.
	if readableDestroyed(rs) {
		return nil
	}
	if !isReading(rs) {
		rs.read()
	}

	data, ok := <-rs.state.buffer

	if ok {
		emitData(rs, data)
	} else {
		destroyReadable(rs)
	}

	return data
}

func destroyReadable(r Readable) {
	if !readableEnded(r) {
		log.Fatal("cannot destroy readable that has not ended")
	}
	r.Emit("readable_end", nil)
}

func (rs *ReadableStream) Pipe(w Writable) Writable {
	attachWritable(rs, w)
	return w
}

func attachWritable(r Readable, w Writable) {
	if readableEnded(r) {
		log.Fatal("cannot call pipe on ended readable stream")
	}
	if readableDestroyed(r) {
		log.Fatal("cannot call pipe on destroyed readable stream")
	}

	state := r.getInternalState()

	if state.dest != nil {
		log.Fatal("cannot call pipe on readable stream, pipe has already been called")
	}

	// Assign writable to source destination
	state.dest = w

	// call resume readable to start flowing data.
	resumeReadable(r)
}

func flow(r Readable) {
	go func() {
		for readableFlowing(r) {
			if r.Read() == nil {
				return
			}
		}
	}()
}

func pauseReadable(r Readable) {
	state := r.getInternalState()
	if state.mode == ReadableFlowing {
		state.mode = ReadableNotFlowing
	}
}

// resumeReadable attempts to restart the stream if it is not
// flowing and not destroyed.
func resumeReadable(r Readable) {
	// if state is destroyed then no point in continuing
	if readableDestroyed(r) {
		return
	}

	if !readableFlowing(r) {
		state := r.getInternalState()
		// If previous value was not already flowing.
		if atomic.SwapUint32(&state.mode, ReadableFlowing) != ReadableFlowing {
			flow(r)
		}
	}
}
