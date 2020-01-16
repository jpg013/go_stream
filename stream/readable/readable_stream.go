package readable

import (
	"container/list"
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
	if state.destroyed || state.reading {
		return
	}

	// We keep reading while we are flowing or until buffer is full
	if readableFlowing(rs) || state.buffer.Len() < state.highWaterMark {
		go rs.Read()
	}
}

func readableFlowing(rs *Stream) bool {
	return atomic.LoadInt32(&rs.state.mode) == ReadableFlowing
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(rs *Stream, data types.Chunk) {
	state := rs.state

	if readableDestroyed(rs) {
		panic("destroyed")
	}

	if readableEnded(rs) {
		panic("push after EOF")
	}

	state.reading = false

	// read chunk is nil, signals end of stream.
	if data == nil {
		onEndOfChunk(rs)
	} else {
		// If flowing and no data buffered then we can skip buffering
		// altogether and just emit the data
		if readableFlowing(rs) &&
			state.buffer.Len() == 0 {
			emitData(rs, data)
		} else {
			// Add data to the buffer for later
			state.buffer.PushBack(data)
		}
	}

	maybeReadMore(rs)
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
	if state.destroyed {
		return
	}

	if !state.destroyed &&
		!readableFlowing(rs) {
		prev := atomic.SwapInt32(&state.mode, ReadableFlowing)

		// If previous value was not already flowing.
		if prev != ReadableFlowing {
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

func emitReadable(rs *Stream) {
	state := rs.state

	if !readableDestroyed(rs) && (state.buffer.Len() > 0 || readableEnded(rs)) {
		rs.Emit("readable", nil)
	}

	flow(rs)
}

func flow(rs *Stream) {
	go func() {
		for readableFlowing(rs) {
			if rs.Read() == nil {
				break
			}
		}
	}()
}

func emitData(rs *Stream, data types.Chunk) {
	state := rs.state
	// increment the counter and emit the data
	atomic.AddInt32(&state.pendingData, 1)
	rs.Emit("data", data)
}

func endReadable(rs *Stream) {
	state := rs.state

	if atomic.LoadInt32(&state.pendingData) > 0 {
		panic("cannot end stream while pending data")
	}

	state.ended = true
	rs.Emit("end", nil)
}

func readableDestroyed(rs *Stream) bool {
	rs.mux.RLock()
	defer rs.mux.RUnlock()
	return rs.state.destroyed == true
}

func readableEnded(rs *Stream) bool {
	rs.mux.RLock()
	defer rs.mux.RUnlock()
	return rs.state.ended == true
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

	bufLen := state.buffer.Len()

	// If we've ended, and we're now clear, then exit.
	if readableEnded(rs) && bufLen == 0 {
		if atomic.LoadInt32(&state.pendingData) == 0 {
			// All pending data has flushed to the writer
			endReadable(rs)
		} else {
			emitReadable(rs)
		}
		return nil
	}

	// Always try to read more if buffer is not full
	doRead := bufLen == 0 || bufLen < state.highWaterMark

	// However, if we are already ready, or ended or destroyed then don't read more.
	if state.reading || readableEnded(rs) || readableDestroyed(rs) {
		doRead = false
	}

	if doRead {
		// Set the reading flag to true and read
		state.reading = true
		rs.read()
	}

	data := fromBuffer(state.buffer)

	if data != nil {
		emitData(rs, data)
	}

	return data
}

func fromBuffer(buf *list.List) types.Chunk {
	if buf.Len() == 0 {
		return nil
	}

	val := buf.Front()

	if val == nil {
		return nil
	}

	buf.Remove(val)

	return val.Value
}

// Pipe method is the main way to start the readable flowing data.
func (rs *Stream) Pipe(w types.Writable) types.Writable {
	if readableEnded(rs) {
		panic("cannot call pipe on ended readable stream")
	}

	if readableDestroyed(rs) {
		panic("cannot call pipe on destroyed readable stream")
	}

	// Lock while we check the destination
	rs.mux.Lock()
	if rs.dest != nil {
		panic("cannot call pipe on readable stream, pipe has already been called")
	}

	// Assign writable to source destination
	rs.dest = w
	rs.mux.Unlock()

	// Flag indicating whether the readable stream has been cleaned up
	cleanedUp := false
	// Flag indicating whether the onDrain event has been added. It would
	// feel a lot more clean to add to do .Once("drain") whenever a drain
	// event is needed but also more expensive and slower.
	onDrain := false

	rs.On("data", func(evt types.Event) {
		ret := rs.dest.Write(evt.Data)

		if !ret {
			// it is possible to get in a permanent state of "paused", when the
			// dest has unpiped from the source.
			if !cleanedUp {
				pauseReadable(rs)
			}

			if !onDrain {
				rs.dest.On("drain", func(evt types.Event) {
					resumeReadable(rs)
				})
			}
		}
		// After each data event, decrement the pendingData count by 1.
		atomic.AddInt32(&rs.state.pendingData, -1)
	})

	rs.Once("end", func(evt types.Event) {
		if readableDestroyed(rs) {
			panic("stream already destroyed")
		}

		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)

		rs.mux.Lock()
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

func (rs *Stream) push(chunk types.Chunk) {
	readableAddChunk(rs, chunk)
}

// NewReadableStream returns a new readable stream instance
func NewReadableStream(conf *Config) (types.Readable, error) {
	if conf.generator == nil {
		return nil, errors.New("cannot create readable stream with nil generator")
	}

	state := NewReadableState()
	gen := conf.generator
	rs := &Stream{
		state:      state,
		doneChan:   make(chan struct{}),
		StreamType: types.ReadableType,
	}

	// Call init of the event emitter
	rs.Emitter.Init()

	// Internal read function
	rs.read = func() {
		if !state.reading {
			panic("cannot read from data source, reading flag not set")
		}

		// Get the next chunk
		chunk, err := gen.Next()

		if err != nil {
			panic(err.Error())
		}

		// add the chunk and "unset" the reading flag
		rs.push(chunk)
	}

	return rs, nil
}
