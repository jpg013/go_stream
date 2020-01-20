package writable

import (
	"errors"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

// Pipe function
func (ws *Stream) Pipe(w types.Writable) types.Writable {
	panic("cannot call pipe on writable stream")
}

func writableBufferChunk(ws *Stream, data types.Chunk) {
	ws.mux.Lock()
	state := ws.state

	state.buffer.PushBack(data)
	atomic.AddInt32(&state.length, 1)

	if getLength(ws) >= state.highWaterMark && !state.draining {
		state.draining = true
	}
	ws.mux.Unlock()
}

func requestWrite(ws *Stream) bool {
	state := ws.state
	ws.mux.Lock()
	defer ws.mux.Unlock()

	if state.writing {
		return false
	}

	state.writing = true
	return state.writing
}

func emitWrite(ws *Stream, chunk types.Chunk) {

}

// If we're already writing something, then just put this
// in the queue, and wait our turn. Otherwise, call doWrite
// If we return false, then we need a drain event, so set that flag.
func writeOrBuffer(ws *Stream, chunk types.Chunk) {
	// If there is not data buffered, and no write requested,
	// then we can simply call emitWritable, and continue.
	if requestWrite(ws) {
		go ws.Emit("write", chunk)
	} else {
		writableBufferChunk(ws, chunk)
	}
}

func getLength(ws *Stream) int {
	return int(atomic.LoadInt32(&ws.state.length))
}

func canWriteMore(ws *Stream) bool {
	return !writableEnded(ws) && getLength(ws) < ws.state.highWaterMark
}

func maybeWriteMore(ws *Stream) {
	// Any data in the buffer?
	if getLength(ws) == 0 {
		return
	}

	if requestWrite(ws) {
		chunk := fromBuffer(ws)

		if chunk != nil {
			go ws.Emit("write", chunk)
		}
	}
}

func onEndOfChunk(ws *Stream) {
	if ws.state.ended {
		panic("Cannot call onEndOfChunk, already ended!")
	}
	ws.mux.Lock()
	ws.state.ended = true
	ws.mux.Unlock()
}

func (ws *Stream) Write(data types.Chunk) bool {
	state := ws.state

	if state.destroyed {
		panic("Error stream destroyed")
	}

	// Similar to a readable stream, a nil chunk indicates end of stream.
	if data == nil {
		onEndOfChunk(ws)
		return false
	}

	writeOrBuffer(ws, data)

	return canWriteMore(ws)
}

func (ws *Stream) Done() <-chan struct{} {
	return ws.doneChan
}

func NewWritableStream(conf *Config) (types.Writable, error) {
	ws := &Stream{
		state:    NewWritableState(),
		Type:     types.WritableType,
		doneChan: make(chan struct{}),
	}

	out := conf.Out

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

		ws.mux.Lock()
		out.Close()
		state.destroyed = true

		// set draining to false if draining
		if state.draining {
			state.draining = false
		}
		ws.mux.Unlock()
		close(ws.doneChan)
	})

	ws.On("write", func(evt types.Event) {
		state := ws.state

		if !state.writing {
			panic("What the fuck")
		}

		// Call output.Write with data
		err := out.Write(evt.Data)

		if err != nil {
			panic(err.Error())
		}

		afterWrite(ws)
	})

	return ws, nil
}

func writableEnded(ws *Stream) bool {
	return ws.state.ended == true
}

func writableDestroyed(ws *Stream) bool {
	return ws.state.destroyed == true
}

func endWritable(rs *Stream) {
	state := rs.state
	rs.mux.Lock()

	if state.writing {
		panic("cannot end writable stream while writing")
	}

	if !state.ended {
		panic("writable not ended")
	}

	state.ended = true
	go rs.Emit("end", nil)
	rs.mux.Unlock()
}

func afterWrite(ws *Stream) {
	state := ws.state

	ws.mux.Lock()
	state.writing = false
	ws.mux.Unlock()

	// Emit drain event if needed
	if getLength(ws) == 0 && state.draining && !writableEnded(ws) {
		go ws.Emit("drain", nil)
		state.draining = false
	}

	ws.mux.RLock()
	shouldEnd := writableEnded(ws) && getLength(ws) == 0 && !state.writing
	ws.mux.RUnlock()

	if shouldEnd {
		endWritable(ws)
	}

	go maybeWriteMore(ws)
}
