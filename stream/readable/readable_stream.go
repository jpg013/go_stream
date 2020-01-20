package readable

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
func maybeReadMore(rs *Stream) {
	state := rs.state

	// If existing read then don't try to read more.
	if readableDestroyed(rs) || readableDone(rs) {
		return
	}

	// We keep reading while we are flowing or until buffer is full
	if readableFlowing(rs) || getLength(rs) < state.highWaterMark {
		go rs.Read()
	}
}

func readableFlowing(rs *Stream) bool {
	return atomic.LoadUint32(&rs.state.mode) == ReadableFlowing
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(rs *Stream, data types.Chunk) {
	if readableDestroyed(rs) {
		panic("destroyed")
	}

	if readableEnded(rs) {
		panic("push after EOF")
	}

	// read chunk is nil, signals end of stream.
	if data == nil {
		onEndOfChunk(rs)
	} else {
		// If flowing and no data buffered then we can skip buffering
		// altogether and just emit the data
		if readableFlowing(rs) &&
			getLength(rs) == 0 &&
			!awaitingDrain(rs) {
			emitData(rs, data)
		} else {
			// Add data to the buffer for later
			readableBufferChunk(rs, data)
		}
	}
	maybeReadMore(rs)
}

func readableBufferChunk(rs *Stream, data types.Chunk) {
	state := rs.state

	state.buffer.PushBack(data)
	atomic.AddInt32(&state.length, 1)
}

func pauseReadable(rs *Stream) {
	if rs.state.mode == ReadableFlowing {
		rs.state.mode = ReadableNotFlowing
	}
}

// resumeReadable attempts to restart the stream if it is not
// flowing and not destroyed.
func resumeReadable(rs *Stream) {
	state := rs.state

	// if state is destroyed then no point in continuing
	if readableDestroyed(rs) {
		return
	}

	if !readableFlowing(rs) {
		// If previous value was not already flowing.
		if atomic.SwapUint32(&state.mode, ReadableFlowing) != ReadableFlowing {
			flow(rs)
		}
	}
}

// Set the ended flag
func onEndOfChunk(rs *Stream) {
	if rs.state.ended {
		panic("stream is already ended")
	}
	rs.state.ended = true
}

func flow(rs *Stream) {
	go func() {
		for readableFlowing(rs) {
			if rs.Read() == nil {
				return
			}
		}
	}()
}

func emitData(rs *Stream, data types.Chunk) {
	state := rs.state
	// increment the counter and emit the data
	atomic.AddInt32(&state.pendingReads, 1)
	rs.Emit("data", data)
}

func endReadable(rs *Stream) {
	state := rs.state

	if hasPendingReads(rs) {
		panic("cannot end stream while pending data")
	}

	state.ended = true
	go rs.Emit("end", nil)
}

func readableDestroyed(rs *Stream) bool {
	return rs.state.destroyed == true
}

func readableEnded(rs *Stream) bool {
	return rs.state.ended == true
}

func isReading(rs *Stream) bool {
	return atomic.LoadUint32(&rs.state.reading) == 1
}

func hasPendingReads(rs *Stream) bool {
	return atomic.LoadInt32(&rs.state.pendingReads) > 0
}

func awaitingDrain(rs *Stream) bool {
	return atomic.LoadUint32(&rs.state.awaitDrainWriters) == 1
}

func readableDone(rs *Stream) bool {
	if readableEnded(rs) &&
		!isReading(rs) &&
		getLength(rs) == 0 &&
		!hasPendingReads(rs) {
		return true
	}
	return false
}

func getLength(rs *Stream) int {
	return int(atomic.LoadInt32(&rs.state.length))
}

// Read is the main method to call in order to either read some
// data from the generator and / or emit a readable event to
// write some data to the destination.
func (rs *Stream) Read() types.Chunk {
	state := rs.state

	// if stream is destroyed then there is no point.
	if readableDestroyed(rs) {
		return nil
	}

	// If we've ended, and we're now clear, then exit.
	if readableDone(rs) {
		// All pending data has flushed to the writer
		endReadable(rs)
		return nil
	}

	bufLen := getLength(rs)

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
		data = fromBuffer(rs)

		if data != nil {
			emitData(rs, data)
		}
	}

	return data
}

func fromBuffer(rs *Stream) types.Chunk {
	state := rs.state

	if getLength(rs) == 0 {
		return nil
	}

	val := state.buffer.Front()

	if val == nil {
		return val
	}

	state.buffer.Remove(val)

	if atomic.AddInt32(&state.length, -1) < 0 {
		panic("cannot have negative length")
	}

	return val.Value
}

// Pipe method is the main way to start the readable flowing data.
func (rs *Stream) Pipe(w types.Writable) types.Writable {
	rs.mux.Lock()
	defer rs.mux.Unlock()

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

	// Flag indicating whether the readable stream has been cleaned up
	// cleanedUp := false

	rs.On("data", func(evt types.Event) {
		ret := rs.dest.Write(evt.Data)
		if !ret {
			// it is possible to get in a permanent state of "paused", when the
			// dest has unpiped from the source.
			pauseReadable(rs)

			if atomic.CompareAndSwapUint32(&rs.state.awaitDrainWriters, 0, 1) {
				rs.dest.Once("drain", func(evt types.Event) {
					atomic.CompareAndSwapUint32(&rs.state.awaitDrainWriters, 1, 0)
					resumeReadable(rs)
				})
			}
		}

		// After each data event, decrement the pendingReads count by 1.
		if atomic.AddInt32(&rs.state.pendingReads, -1) < 0 {
			panic("Cannot have negative pending reads")
		}
	})

	rs.Once("end", func(evt types.Event) {
		rs.mux.Lock()

		if readableDestroyed(rs) {
			return
		}

		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)
		rs.state.destroyed = true
		rs.mux.Unlock()

		// Call close on doneChan to send signal to the channel
		close(rs.doneChan)
	})

	// call resume readable to start flowing data.
	resumeReadable(rs)

	// return the writable stream
	return w
}

// Done returns the readable doneChan
func (rs *Stream) Done() <-chan struct{} {
	return rs.doneChan
}

// NewReadableStream returns a new readable stream instance
func NewReadableStream(conf *Config) (types.Readable, error) {
	if conf.Generator == nil {
		return nil, errors.New("readable stram requires data source")
	}

	state := NewReadableState()
	gen := conf.Generator
	rs := &Stream{
		state:      state,
		doneChan:   make(chan struct{}),
		StreamType: types.ReadableType,
	}

	// Internal read function
	rs.read = func() {
		if !atomic.CompareAndSwapUint32(&state.reading, 0, 1) {
			return
		}

		for howMuchToRead(rs) > 0 {
			// Get the next chunk
			chunk, err := gen.Next()

			if err != nil {
				panic(err.Error())
			}

			// add the chunk and "unset" the reading flag
			readableAddChunk(rs, chunk)

			// break loop on nil chunk
			if chunk == nil {
				break
			}
		}

		if !atomic.CompareAndSwapUint32(&state.reading, 1, 0) {
			panic("Could not unset reading flag")
		}
	}

	return rs, nil
}

func howMuchToRead(rs *Stream) int {
	state := rs.state

	if getLength(rs) < state.highWaterMark && !state.ended {
		return 1
	}

	return 0
}
