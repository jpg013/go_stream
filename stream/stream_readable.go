package stream

import (
	"errors"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

// From a high level, a readable stream can be thought of as having two separate, yet competing
// processes. On one side, the readable stream is attempting to consume as much data as possible
// from a data generator, and on the other the readable stream is trying to write the data to an
// underlying writable stream. These interactions can result in 3 scenarios:
// 1) The readable stream is both consuming data and writing data in perfect harmony. In other words,
// for each chunk of data consumed from the data generator, the readable stream is also able to
// write an equal chunk of data to the underlying destination. This results in the readable stream
// never having to wait for a writable "drain" event, nor the writable stream ever having to "wait"
// for a readable event.
// 2) The readable stream is consuming data faster than can be written to writable stream. This will
// cause the internal buffer to fill up and overflow, at which point the readable stream will need
// to pause data consumption until a writable "drain" event occurs.
// 3) The writable stream processes data faster than the readable stream is able to produce. In this case
// we set a "needReadable" flag in order to know to emit data as soon as we have some available.
func maybeReadMore(r Readable) {
	rs := r.GetReadableState()

	// If existing read then don't try to read more.
	if readableDestroyed(r) || readableFinished(r) {
		return
	}

	len := int(atomic.LoadInt32(&rs.length))
	// We keep reading while we are flowing or until buffer is full
	if readableFlowing(r) || len < rs.highWaterMark {
		go r.Read()
	}
}

func readableFlowing(r Readable) bool {
	state := r.GetReadableState()

	return atomic.LoadUint32(&state.mode) == ReadableFlowing
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(r Readable, data types.Chunk) {
	if readableDestroyed(r) {
		panic("destroyed")
	}

	if readableEnded(r) {
		panic("push after EOF")
	}

	// read chunk is nil, signals end of stream.
	if data == nil {
		readableEndOfChunk(r)
	} else {
		rs := r.GetReadableState()
		len := atomic.LoadInt32(&rs.length)

		doRead := readableFlowing(r) && len == 0 && !awaitingDrain(r)

		// If flowing and no da	ta buffered then we can skip buffering
		// altogether and just emit the data
		if doRead {
			emitData(r, data)
		} else {
			// Add data to the buffer for later
			readableBufferChunk(r, data)
		}
	}
	maybeReadMore(r)
}

func readableBufferChunk(r Readable, data types.Chunk) {
	rs := r.GetReadableState()

	rs.buffer.PushBack(data)
	atomic.AddInt32(&rs.length, 1)
}

func pauseReadable(r Readable) {
	rs := r.GetReadableState()
	if rs.mode == ReadableFlowing {
		rs.mode = ReadableNotFlowing
	}
}

// resumeReadable attempts to restart the stream if it is not
// flowing and not destroyed.
func resumeReadable(r Readable) {
	rs := r.GetReadableState()

	// if state is destroyed then no point in continuing
	if readableDestroyed(r) {
		return
	}

	if !readableFlowing(r) {
		// If previous value was not already flowing.
		if atomic.SwapUint32(&rs.mode, ReadableFlowing) != ReadableFlowing {
			flow(r)
		}
	}
}

// Set the ended flag
func readableEndOfChunk(r Readable) {
	if readableEnded(r) {
		panic("stream is already ended")
	}
	state := r.GetReadableState()
	state.ended = true
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

func emitData(r Readable, data types.Chunk) {
	rs := r.GetReadableState()

	// increment the counter and emit the data
	atomic.AddInt32(&rs.pendingReads, 1)
	r.Emit("data", data)
}

func endReadable(r Readable) {
	rs := r.GetReadableState()

	if hasPendingReads(r) {
		panic("cannot end stream while pending data")
	}

	rs.ended = true
	r.Emit("readable_end", nil)
}

func readableDestroyed(r Readable) bool {
	state := r.GetReadableState()
	return state.destroyed
}

func readableEnded(r Readable) bool {
	state := r.GetReadableState()
	return state.ended
}

func isReading(r Readable) bool {
	state := r.GetReadableState()
	return atomic.LoadUint32(&state.reading) == 1
}

func hasPendingReads(r Readable) bool {
	rs := r.GetReadableState()
	return atomic.LoadInt32(&rs.pendingReads) > 0
}

func awaitingDrain(r Readable) bool {
	state := r.GetReadableState()

	return atomic.LoadUint32(&state.awaitDrainWriters) == 1
}

func readableFinished(r Readable) bool {
	rs := r.GetReadableState()
	len := atomic.LoadInt32(&rs.length)

	if readableEnded(r) &&
		!isReading(r) &&
		len == 0 &&
		!hasPendingReads(r) {
		return true
	}
	return false
}

// Read is the main method to call in order to either read some
// data from the generator and / or emit a readable event to
// write some data to the destination.
func (rs *ReadableStream) Read() types.Chunk {
	state := rs.state

	// if stream is destroyed then there is no point.
	if readableDestroyed(rs) {
		return nil
	}

	// If we've ended, and we're now clear, then exit.
	if readableFinished(rs) {
		// All pending data has flushed to the writer
		endReadable(rs)
		return nil
	}

	bufLen := int(atomic.LoadInt32(&rs.state.length))

	// Always try to read more if buffer is not full
	doRead := bufLen == 0 || bufLen < state.highWaterMark

	// However, if we are already ready, or ended or destroyed then don't read more.
	if isReading(rs) || readableEnded(rs) || readableDestroyed(rs) {
		doRead = false
	}

	if doRead {
		rs.read()
	}

	var data types.Chunk

	if !awaitingDrain(rs) {
		data = fromReadableBuffer(state)

		if data != nil {
			emitData(rs, data)
		}
	}

	return data
}

func attachWritable(r Readable, w Writable) Writable {
	if readableEnded(r) {
		panic("cannot call pipe on ended readable stream")
	}

	if readableDestroyed(r) {
		panic("cannot call pipe on destroyed readable stream")
	}

	rs := r.GetReadableState()

	if rs.dest != nil {
		panic("cannot call pipe on readable stream, pipe has already been called")
	}

	// Assign writable to source destination
	rs.dest = w

	// call resume readable to start flowing data.
	resumeReadable(r)

	// return the writable stream
	return w
}

// Pipe method is the main way to start the readable flowing data.
func (rs *ReadableStream) Pipe(w Writable) Writable {
	return attachWritable(rs, w)
}

// Done returns the readable doneChan
func (rs *ReadableStream) Done() <-chan struct{} {
	return rs.doneChan
}

func (rs *ReadableStream) GetReadableState() *ReadableState {
	return rs.state
}

// NewReadableStream returns a new readable stream instance
func NewReadableStream(conf *Config) (Readable, error) {
	gen := conf.Readable.Generator

	if gen == nil {
		return nil, errors.New("readable stram requires data source")
	}

	stream := &ReadableStream{
		state:      NewReadableState(),
		doneChan:   make(chan struct{}),
		StreamType: ReadableType,
	}

	// Internal read function
	stream.read = func() {
		rs := stream.state
		if !atomic.CompareAndSwapUint32(&rs.reading, 0, 1) {
			return
		}

		for howMuchToRead(rs) > 0 {
			// Get the next chunk
			chunk, err := gen.Next()

			if err != nil {
				panic(err.Error())
			}

			// add the chunk and "unset" the reading flag
			readableAddChunk(stream, chunk)

			// break loop on nil chunk
			if chunk == nil {
				break
			}
		}

		if !atomic.CompareAndSwapUint32(&rs.reading, 1, 0) {
			panic("Could not unset reading flag")
		}
	}

	stream.On("data", func(evt types.Event) {
		rs := stream.state
		ret := rs.dest.Write(evt.Data)

		if !ret {
			// it is possible to get in a permanent state of "paused", when the
			// dest has unpiped from the source.
			pauseReadable(stream)

			if atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 0, 1) {
				rs.dest.Once("drain", func(evt types.Event) {
					atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 1, 0)
					resumeReadable(stream)
				})
			}
		}

		// After each data event, decrement the pendingReads count by 1.
		if atomic.AddInt32(&rs.pendingReads, -1) < 0 {
			panic("Cannot have negative pending reads")
		}
	})

	stream.Once("readable_end", func(evt types.Event) {
		if readableDestroyed(stream) {
			return
		}

		rs := stream.state

		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)
		rs.destroyed = true

		// Call close on doneChan to send signal to the channel
		close(stream.doneChan)
	})

	return stream, nil
}

func howMuchToRead(state *ReadableState) int {
	len := int(atomic.LoadInt32(&state.length))

	if len < state.highWaterMark && !state.ended {
		return 1
	}

	return 0
}
