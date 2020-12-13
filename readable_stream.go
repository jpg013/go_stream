package main

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/types"
)

const (
	// ReadableFlowing mode reads from the underlying system automatically
	// and provides to an application as quickly as possible.
	ReadableFlowing uint32 = 1
	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
	ReadableNotFlowing uint32 = 2
	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
	ReadableNull uint32 = 0
)

type Readable interface {
	Stream
	Read() types.Chunk
}

type ReadableStream struct {
	// embedded event emitter
	emitter.Emitter

	mux               sync.RWMutex
	buffer            chan (types.Chunk)
	mode              uint32
	highWaterMark     uint
	destroyed         bool
	pendingReads      int32
	awaitDrainWriters uint32
	doneCh            chan (struct{})

	Type StreamType

	ended bool

	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest Writable

	// bit determining whether the stream is currently reading or not
	reading uint32

	// What is this???
	length int64

	// internal read method that can be overwritten
	read func()
}

func acquireRead(rs *ReadableStream) bool {
	return atomic.CompareAndSwapUint32(&rs.reading, 0, 1)
}

func releaseRead(rs *ReadableStream) bool {
	return atomic.CompareAndSwapUint32(&rs.reading, 1, 0)
}

func NewReadable(conf ReadableConfig) (r Readable, err error) {
	rs := &ReadableStream{
		highWaterMark: conf.highWaterMark,
		read:          conf.read,
		mode:          ReadableNull,
		buffer:        make(chan types.Chunk, 100),
		doneCh:        make(chan struct{}),
		reading:       0,
	}

	if rs.read == nil {
		err = applyRead(rs, conf.generator)
	}

	rs.On("data", func(evt types.Event) {
		if !rs.dest.Write(evt.Data) {
			// it is possible to get in a permanent state of "paused", when the
			// dest has unpiped from the source.
			pauseReadable(rs)

			if atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 0, 1) {
				rs.dest.Once("drain", func(evt types.Event) {
					atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 1, 0)
					resumeReadable(rs)
				})
			}
		}

		// After each data event, decrement the pendingReads count by 1.
		if atomic.AddInt32(&rs.pendingReads, -1) < 0 {
			log.Fatal("Cannot have negative pending reads")
		}
	})

	rs.Once("readable_end", func(evt types.Event) {
		if readableDestroyed(rs) {
			return
		}

		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)
		rs.destroyed = true
		// Call close on doneChan to send signal to the channel
		close(rs.doneCh)
	})

	return rs, err
}

func howMuchToRead(rs *ReadableStream) int {
	len := uint(atomic.LoadInt64(&rs.length))

	if len < rs.highWaterMark && !readableEnded(rs) {
		return 1
	}

	return 0
}

func applyRead(rs *ReadableStream, gen generators.Type) error {
	if gen == nil {
		return errors.New("readable stream generator is requred")
	}

	rs.read = func() {
		if !acquireRead(rs) {
			return
		}

		go func() {
			for howMuchToRead(rs) > 0 {
				// Get the next chunk
				chunk, err := gen.Next()

				if err != nil {
					log.Fatal(err.Error())
				}

				// add the chunk and "unset" the reading flag
				readableAddChunk(rs, chunk)

				if chunk == nil {
					readableEndOfChunk(rs)
				}
			}

			if !releaseRead(rs) {
				log.Fatal("Could not unset reading flag")
			}

			// Check if we should read more
			maybeReadMore(rs)
		}()
	}

	return nil
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(rs *ReadableStream, data types.Chunk) {
	if readableDestroyed(rs) {
		log.Fatal("readable is destroyed")
	}

	if readableEnded(rs) {
		log.Fatal("push after readable ended")
	}

	// Add chunk to buffer
	readableBufferChunk(rs, data)
}

func readableDestroyed(rs *ReadableStream) bool {
	return rs.destroyed
}

func readableEnded(rs *ReadableStream) bool {
	return rs.ended
}

// Set the ended flag
func readableEndOfChunk(rs *ReadableStream) {
	if readableEnded(rs) {
		log.Fatal("stream is already ended")
	}
	rs.ended = true
	close(rs.buffer)
}

func readableFlowing(rs *ReadableStream) bool {
	return atomic.LoadUint32(&rs.mode) == ReadableFlowing
}

func emitData(rs *ReadableStream, data types.Chunk) {
	// increment the counter and emit the data
	atomic.AddInt32(&rs.pendingReads, 1)
	rs.Emit("data", data)
}

func maybeReadMore(rs *ReadableStream) {
	// If existing read then don't try to read more.
	if readableDestroyed(rs) || readableEnded(rs) {
		return
	}

	len := uint(atomic.LoadInt64(&rs.length))

	// We keep reading while we are flowing or until buffer is full
	if readableFlowing(rs) && len < rs.highWaterMark {
		rs.Read()
	}
}

func (rs *ReadableStream) Done() <-chan struct{} {
	return rs.doneCh
}

// Read is the main method to call in order to either read some
// data from the generator and / or emit a readable event to
// write some data to the destination.
func (rs *ReadableStream) Read() types.Chunk {
	// if stream is destroyed then there is no point.
	if readableDestroyed(rs) {
		return nil
	}

	bufLen := uint(atomic.LoadInt64(&rs.length))

	// Always try to read more if buffer is not full
	doRead := bufLen == 0 || bufLen < rs.highWaterMark

	// However, if we are already ready, or ended or destroyed then don't read more.
	if isReading(rs) || readableEnded(rs) || readableDestroyed(rs) {
		doRead = false
	}

	if doRead {
		rs.read()
	}

	var data types.Chunk
	data = <-rs.buffer

	// increment counter
	atomic.AddInt64(&rs.length, -1)

	if data != nil {
		emitData(rs, data)
	}

	// If we've ended, and we're now clear, then exit.
	if readableFinished(rs) {
		// All pending data has flushed to the writer
		destroyReadable(rs)
	}

	return data
}

func readableBufferChunk(rs *ReadableStream, data types.Chunk) {
	// push data onto buffer
	rs.buffer <- data

	// increment counter
	atomic.AddInt64(&rs.length, 1)
}

func readableFinished(rs *ReadableStream) bool {
	if readableEnded(rs) &&
		!isReading(rs) &&
		atomic.LoadInt64(&rs.length) == 0 &&
		!hasPendingReads(rs) {
		return true
	}
	return false
}

func hasPendingReads(rs *ReadableStream) bool {
	return atomic.LoadInt32(&rs.pendingReads) > 0
}

func isReading(rs *ReadableStream) bool {
	return atomic.LoadUint32(&rs.reading) == 1
}

func destroyReadable(rs *ReadableStream) {
	if !readableEnded(rs) {
		log.Fatal("cannot destroy readable that has not ended")
	}

	if hasPendingReads(rs) {
		log.Fatal("cannot end stream while pending data")
	}

	rs.Emit("readable_end", nil)
}

func (rs *ReadableStream) Pipe(w Writable) Writable {
	attachWritable(rs, w)

	return w
}

func attachWritable(rs *ReadableStream, w Writable) {
	if readableEnded(rs) {
		panic("cannot call pipe on ended readable stream")
	}

	if readableDestroyed(rs) {
		panic("cannot call pipe on destroyed readable stream")
	}

	if rs.dest != nil {
		panic("cannot call pipe on readable stream, pipe has already been called")
	}

	// Assign writable to source destination
	rs.dest = w

	// call resume readable to start flowing data.
	resumeReadable(rs)
}

func flow(rs *ReadableStream) {
	go func() {
		for readableFlowing(rs) {
			if rs.Read() == nil {
				return
			}
		}
	}()
}

func pauseReadable(rs *ReadableStream) {
	if rs.mode == ReadableFlowing {
		rs.mode = ReadableNotFlowing
	}
}

// resumeReadable attempts to restart the stream if it is not
// flowing and not destroyed.
func resumeReadable(rs *ReadableStream) {
	// if state is destroyed then no point in continuing
	if readableDestroyed(rs) {
		return
	}

	if !readableFlowing(rs) {
		// If previous value was not already flowing.
		if atomic.SwapUint32(&rs.mode, ReadableFlowing) != ReadableFlowing {
			flow(rs)
		}
	}
}
