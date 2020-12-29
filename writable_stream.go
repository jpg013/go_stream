package main

import (
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/types"
)

type StreamType uint32

type Stream interface {
	types.EventEmitter
	Pipe(Writable) Writable
	Done() <-chan struct{}
	getInternalState() *InternalState
}

type Writable interface {
	Stream
	Write(types.Chunk) bool
}

// Config represents a writable config
type WritableConfig struct {
	highWaterMark uint64
	output        output.Type
}

type WritableStream struct {
	emitter.Emitter
	Type    StreamType
	writing uint32
	state   *InternalState
}

func writableEndOfChunk(w Writable) {
	if writableEnded(w) {
		log.Fatal("Cannot call writableEndOfChunk(), writable_stream already ended")
	}

	if !endWritable(w) {
		log.Fatal("Could not end writable, writable already ended.")
	}
}

func (ws *WritableStream) Write(data types.Chunk) bool {
	if writableEnded(ws) {
		log.Fatal("Cannot write to ended stream")
	}

	if writableDestroyed(ws) {
		log.Fatal("Cannot write to destroyed stream")
	}

	// Similar to a readable stream, a nil chunk indicates end of stream.
	if data == nil {
		writableEndOfChunk(ws)
		return false
	} else {
		writableBufferChunk(ws, data)
	}

	doWrite := !isWriting(ws) && !writableEnded(ws)

	if doWrite {
		write(ws)
	}

	return true
}

func write(w Writable) {
	if !acquireWrite(w) {
		return
	}

	go func() {
		state := w.getInternalState()
		for chunk := range state.buffer {
			w.Emit("write", chunk)
		}
		afterWrite(w)
		if !releaseWrite(w) {
			log.Fatal("Could not unset writing flag")
		}
	}()
}

func (ws *WritableStream) Done() <-chan struct{} {
	return ws.state.doneCh
}

// Pipe function
func (ws *WritableStream) Pipe(w Writable) Writable {
	panic("cannot call pipe on writable stream")
}

func (ws *WritableStream) getInternalState() *InternalState {
	return ws.state
}

func NewWritableStream(conf StreamConfig) (Writable, error) {
	ws := &WritableStream{
		state: NewInternalState(conf),
	}

	out := conf.output

	if out == nil {
		return nil, errors.New("writable stream config requires output type")
	}

	ws.On("write", func(evt types.Event) {
		time.Sleep(10 * time.Millisecond)
		// Call output.Write with data
		err := out.Write(evt.Data)

		if err != nil {
			log.Fatal(err.Error())
		}
	})

	ws.Once("writable_done", func(evt types.Event) {
		if writableDestroyed(ws) {
			log.Fatal("cannot end writable stream, already destroyed")
		}

		if !writableEnded(ws) {
			log.Fatal("\"writable_done\" event emitted before end of stream")
		}

		// Close the output
		out.Close()

		if !atomic.CompareAndSwapInt32(&ws.state.writableDestroyed, 0, 1) {
			log.Fatal("could not destroy writable - already destroyed")
		}

		// unset if draining
		if isDraining(ws) {
			atomic.CompareAndSwapUint32(&ws.state.draining, 1, 0)
		}

		// Send message to done channel
		close(ws.state.doneCh)
	})

	return ws, nil
}

func writableEnded(w Writable) bool {
	state := w.getInternalState()
	return atomic.LoadInt32(&state.writableEnded) == 1
}

func writableBufferChunk(w Writable, chunk types.Chunk) {
	state := w.getInternalState()
	state.buffer <- chunk
}

func isDraining(w Writable) bool {
	state := w.getInternalState()
	return atomic.LoadUint32(&state.draining) == 1
}

func isWriting(w Writable) bool {
	state := w.getInternalState()
	return atomic.LoadUint32(&state.writing) == uint32(1)
}

func acquireWrite(w Writable) bool {
	state := w.getInternalState()
	return atomic.CompareAndSwapUint32(&state.writing, 0, 1)
}

func releaseWrite(w Writable) bool {
	state := w.getInternalState()
	return atomic.CompareAndSwapUint32(&state.writing, 1, 0)
}

func endWritable(w Writable) bool {
	state := w.getInternalState()
	if atomic.CompareAndSwapInt32(&state.writableEnded, 0, 1) {
		close(state.buffer)
		return true
	}
	return false
}

func writableDestroyed(w Writable) bool {
	state := w.getInternalState()
	return atomic.LoadInt32(&state.writableDestroyed) == 1
}

func afterWrite(w Writable) {
	w.Emit("writable_done", nil)
}
