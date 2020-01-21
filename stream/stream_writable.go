package stream

import (
	"errors"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

// Pipe function
func (ws *WritableStream) Pipe(w Writable) Writable {
	panic("cannot call pipe on writable stream")
}

func writeAvailable(w Writable) bool {
	state := w.GetWriteState()

	return atomic.CompareAndSwapUint32(&state.writing, 0, 1)
}

// If we're already writing something, then just put this
// in the queue, and wait our turn. Otherwise, call doWrite
// If we return false, then we need a drain event, so set that flag.
func writeOrBuffer(w Writable, chunk types.Chunk) {
	// If there is not data buffered, and no write requested,
	// then we can simply call emitWritable, and continue.
	if writeAvailable(w) {
		go w.Emit("write", chunk)
	} else {
		writableBufferChunk(w, chunk)
	}
}

func canWriteMore(w Writable) bool {
	state := w.GetWriteState()
	len := int(atomic.LoadInt32(&state.length))

	return !writableEnded(w) && len < state.highWaterMark
}

func maybeWriteMore(w Writable) {
	state := w.GetWriteState()

	// Any data in the buffer?
	len := atomic.LoadInt32(&state.length)

	if len == 0 {
		return
	}

	if writeAvailable(w) {
		chunk := shiftWritableBuffer(w)

		if chunk == nil {
			panic("received nil chunk from buffer")
		}

		w.Emit("write", chunk)
	}
}

func writableEndOfChunk(w Writable) {
	state := w.GetWriteState()

	if state.ended {
		panic("Cannot call writableEndOfChunk(), writable_stream already ended")
	}

	state.ended = true
}

func (ws *WritableStream) Write(data types.Chunk) bool {
	state := ws.state

	if state.destroyed {
		panic("Error stream destroyed")
	}

	// Similar to a readable stream, a nil chunk indicates end of stream.
	if data == nil {
		writableEndOfChunk(ws)
		return false
	}

	writeOrBuffer(ws, data)

	return canWriteMore(ws)
}

func (ws *WritableStream) Done() <-chan struct{} {
	return ws.doneChan
}

func (ws *WritableStream) GetWriteState() *WritableState {
	return ws.state
}

func NewWritableStream(conf *Config) (Writable, error) {
	ws := &WritableStream{
		state:    NewWritableState(),
		Type:     WritableType,
		doneChan: make(chan struct{}),
	}

	out := conf.Writable.Out

	if out == nil {
		return nil, errors.New("writable stream requires config with output type")
	}

	ws.Once("end", func(evt types.Event) {
		state := ws.state

		if writableDestroyed(ws) {
			panic("cannot end writable stream, already destroyed")
		}

		if !writableEnded(ws) {
			panic("\"end\" event emitted before end of stream")
		}

		out.Close()
		state.destroyed = true

		// unset if draining
		if atomic.LoadUint32(&state.draining) == 1 {
			atomic.CompareAndSwapUint32(&state.draining, 1, 0)
		}
		close(ws.doneChan)
	})

	ws.On("write", func(evt types.Event) {
		// Call output.Write with data
		err := out.Write(evt.Data)

		if err != nil {
			panic(err.Error())
		}

		afterWrite(ws)
	})

	return ws, nil
}

func writableEnded(w Writable) bool {
	state := w.GetWriteState()

	return state.ended == true
}

func writableDestroyed(ws *WritableStream) bool {
	return ws.state.destroyed == true
}

func endWritable(w Writable) {
	rs := w.GetWriteState()

	writing := atomic.LoadUint32(&rs.writing) > 0

	if writing {
		panic("cannot end writable stream while writing")
	}

	if !rs.ended {
		panic("writable not ended")
	}

	rs.ended = true
	go w.Emit("end", nil)
}

func afterWrite(w Writable) {
	state := w.GetWriteState()

	if !atomic.CompareAndSwapUint32(&state.writing, 1, 0) {
		panic("What in the holiest of fucks")
	}

	len := atomic.LoadInt32(&state.length)
	writing := atomic.LoadUint32(&state.writing) > 0
	needDrain := (len == 0 &&
		atomic.LoadUint32(&state.draining) == 1 &&
		!writableEnded(w))
	needEnd := (writableEnded(w) &&
		len == 0 &&
		!writing)

	// Emit drain event if needed
	if needDrain {
		if atomic.CompareAndSwapUint32(&state.draining, 1, 0) {
			go w.Emit("drain", nil)
		}
	}

	if needEnd {
		endWritable(w)
	}

	go maybeWriteMore(w)
}
