package writable

import (
	"stream/emitter"
	"stream/output"
	"stream/types"
)

// Stream type
type Stream struct {
	emitter *emitter.Emitter
	state   *WritableState
}

// On function
func (ws *Stream) On(topic string, fn types.EventHandler) {
	ws.emitter.On(topic, fn)
}

// Emit function
func (ws *Stream) Emit(topic string, data interface{}) {
	ws.emitter.Emit(topic, data)
}

// Pipe function
func (ws *Stream) Pipe(w Writable) Writable {
	panic("cannot pipe to writable stream")
}

func emitWritable(ws *Stream, data types.Chunk) {
	state := ws.state

	if state.writeRequested {
		panic("cannot call emit writable, write already requested")
	}

	state.writeRequested = true
	ws.Emit("write", data)
}

func writableBufferChunk(w *Stream, data types.Chunk) {
	state := w.state
	state.buffer.PushBack(data)

	if state.buffer.Len() >= state.highWaterMark && !state.draining {
		state.draining = true
	}
}

// If we're already writing something, then just put this
// in the queue, and wait our turn. Otherwise, call doWrite
// If we return false, then we need a drain event, so set that flag.
func writeOrBuffer(w *Stream, chunk types.Chunk) bool {
	state := w.state

	// If there is not data buffered, and no write requested,
	// then we can simply call emitWritable, and continue.
	if !state.writeRequested && state.buffer.Len() == 0 {
		emitWritable(w, chunk)
	} else {
		writableBufferChunk(w, chunk)
	}

	return canWriteMore(state)
}

func canWriteMore(state *WritableState) bool {
	return state.buffer.Len() < state.highWaterMark
}

func writeNext(ws *Stream) {
	state := ws.state
	state.mtx.Lock()
	defer state.mtx.Unlock()

	if state.writeRequested || state.buffer.Len() == 0 || state.destroyed {
		return
	}

	chunk := fromBuffer(ws.state.buffer)

	emitWritable(ws, chunk)
}

func writableEndOfChunk(ws *Stream) {
	ws.state.ended = true
}

func (w *Stream) Write(data types.Chunk) bool {
	state := w.state
	state.mtx.Lock()
	defer state.mtx.Unlock()

	if state.destroyed {
		panic("Error stream destroyed")
	}

	// Similar to a readable stream, a nil chunk indicates end of stream.
	if data == nil {
		writableEndOfChunk(w)
		return false
	}

	return writeOrBuffer(w, data)
}

func NewWritable(out output.Type) Writable {
	w := &Stream{
		state:   NewWritableState(),
		emitter: emitter.NewEmitter(),
	}

	w.On("write", func(evt types.Event) {
		state := w.state
		state.mtx.Lock()
		defer state.mtx.Unlock()

		if !state.writeRequested {
			panic("should not be in write event handler without write requested")
		}

		// Call output.Write with data
		err := out.Write(evt.Data)

		if err != nil {
			panic(err.Error())
		}

		if state.ended && state.buffer.Len() == 0 {
			out.Close()
			state.destroyed = true
			state.draining = false
		} else if state.buffer.Len() == 0 && state.draining && !state.ended {
			w.emitter.Emit("drain", nil)
			state.draining = false
		}

		// unset the write next
		state.writeRequested = false

		go writeNext(w)
	})

	return w
}
