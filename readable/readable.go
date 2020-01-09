package readable

import (
	"container/list"
	"stream/emitter"
	"stream/generators"
	"stream/types"
	"stream/writable"
	"sync"
)

func maybeReadMore(rs *Stream) {
	state := rs.state

	if state.destroyed {
		return
	}

	if isReadableFlowing(rs) {
		go rs.Read()
	}
}

func isReadableFlowing(rs *Stream) bool {
	return rs.state.mode == ReadableFlowing
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(rs *Stream, data types.Chunk) {
	state := rs.state

	// read chunk is nil, signals end of stream.
	if data == nil {
		onEndOfChunk(rs)
	} else {
		if isReadableFlowing(rs) &&
			!state.readRequested &&
			state.buffer.Len() == 0 {
			emitReadable(rs, data)
		} else {
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

func resumeReadable(rs *Stream) {
	state := rs.state

	// if state is destroyed then no point in continuing
	if state.destroyed {
		return
	}

	if rs.state.mode != ReadableFlowing {
		rs.state.mode = ReadableFlowing
		go rs.Read()
	}
}

func fromBuffer(buf *list.List) interface{} {
	val := buf.Front()

	if val == nil {
		return val
	}

	buf.Remove(val)

	return val.Value
}

// Set the ended flag
func onEndOfChunk(rs *Stream) {
	if rs.state.ended {
		panic("stream is already ended")
	}
	rs.state.ended = true
}

type Stream struct {
	emitter *emitter.Emitter
	state   *ReadableState
	dest    writable.Writable
	mux     sync.Mutex
	gen     generators.Type
}

func canReadMore(rs *Stream) bool {
	state := rs.state

	return !state.destroyed &&
		isReadableFlowing(rs) &&
		state.buffer.Len() < state.highWaterMark
}

func emitReadable(rs *Stream, data types.Chunk) {
	state := rs.state

	// if a read has been requested then exit
	if state.readRequested || !isReadableFlowing(rs) || data == nil {
		panic("cannot emit readable")
	}

	// Set read requested to true and emit the data
	state.readRequested = true
	// emit chunk and continue
	rs.Emit("data", data)
}

func (rs *Stream) Read() bool {
	state := rs.state
	state.mtx.Lock()
	defer state.mtx.Unlock()

	// if stream is destroyed then there is not point.
	if state.destroyed {
		panic("stream is destroyed")
	}

	// If we have data in the buffer and we are flowing and a read has not been requested
	// then call emitReadable to write data to the underlying writable stream
	if isReadableFlowing(rs) &&
		state.buffer.Len() > 0 &&
		!state.readRequested {
		emitReadable(rs, fromBuffer(rs.state.buffer))
	}

	// If the readable stream has not ended and can push more data into the buffer the continue
	if !state.ended && state.buffer.Len() < state.highWaterMark {
		chunk, _ := rs.gen.Next()
		readableAddChunk(rs, chunk)
	}

	return canReadMore(rs)
}

func (rs *Stream) On(topic string, fn types.EventHandler) {
	rs.emitter.On(topic, fn)
}

func (rs *Stream) Emit(topic string, data interface{}) {
	rs.emitter.Emit(topic, data)
}

func (rs *Stream) Pipe(w writable.Writable) writable.Writable {
	if rs.state.ended {
		panic("cannot Pipe() to ended readable")
	}

	if rs.state.destroyed {
		panic("cannot Pipe() to destroyed readable")
	}

	if rs.dest != nil {
		panic("Cannot pipe Writable - Readable destination already exists")
	}

	// Assign writable to source destination
	rs.dest = w

	rs.On("data", func(evt types.Event) {
		if !isReadableFlowing(rs) {
			panic("readable stream is not flowing")
		}

		if !rs.dest.Write(evt.Data) {
			pauseReadable(rs)
		}

		// unset the read requested flag and attempt to read more data
		rs.state.readRequested = false

		// Check to see if the readable stream has ended
		if rs.state.ended && rs.state.buffer.Len() == 0 {
			rs.Emit("end", nil)
		} else {
			// Trying reading more if we can
			maybeReadMore(rs)
		}
	})

	rs.On("end", func(evt types.Event) {
		if rs.state.destroyed {
			panic("we have already been here")
		}
		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)
		rs.state.destroyed = true
	})

	rs.On("destroy", func(evt types.Event) {
		// Handle destroy
		if rs.state.destroyed {
			panic("cannot destroy stream, already destroyed")
		}
		rs.state.destroyed = true
	})

	w.On("drain", func(evt types.Event) {
		resumeReadable(rs)
	})

	// Resume reading
	resumeReadable(rs)

	return w
}

func NewReadable(gen generators.Type) Readable {
	r := &Stream{
		emitter: emitter.NewEmitter(),
		state:   NewReadableState(),
		gen:     gen,
	}

	return r
}
