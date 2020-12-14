package main

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/types"
)

type StreamType uint32

type Stream interface {
	types.EventEmitter
	Pipe(Writable) Writable
	Done() <-chan struct{}
}

type Writable interface {
	Stream
	Write(types.Chunk) bool
}

// Config represents a writable config
type WritableConfig struct {
	highWaterMark uint
	output        output.Type
}

type WritableStream struct {
	emitter.Emitter
	Type          StreamType
	doneCh        chan struct{}
	highWaterMark uint
	length        int64
	waiting       int32
	buffer        Buffer
	destroyed     bool
	ended         int32
	draining      uint32
	mux           sync.RWMutex
	writing       uint32
}

func writableEndOfChunk(ws *WritableStream) {
	if writableEnded(ws) {
		log.Fatal("Cannot call writableEndOfChunk(), writable_stream already ended")
	}

	if !endWritable(ws) {
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
		if writableFinished(ws) {
			ws.Emit("writable_done", nil)
		}
		return false
	} else {
		writableBufferChunk(ws, data)
	}

	doWrite := !isWriting(ws) && !writableEnded(ws)

	if doWrite {
		write(ws)
	}

	return canWriteMore(ws)
}

func howMuchToWrite(ws *WritableStream) int {
	return ws.buffer.Len()
}

func write(ws *WritableStream) {
	if !acquireWrite(ws) {
		return
	}

	go func() {
		for howMuchToWrite(ws) > 0 {
			chunk, _ := ws.buffer.Read()
			ws.Emit("write", chunk)
		}
		if !releaseWrite(ws) {
			log.Fatal("Could not unset writing flag")
		}
		maybeWriteMore(ws)
	}()
}

func maybeWriteMore(ws *WritableStream) {
	if !isWriting(ws) && ws.buffer.Len() > 0 {
		write(ws)
	}
}

func (ws *WritableStream) Done() <-chan struct{} {
	return ws.doneCh
}

// Pipe function
func (ws *WritableStream) Pipe(w Writable) Writable {
	panic("cannot call pipe on writable stream")
}

func NewWritableStream(conf WritableConfig) (Writable, error) {
	ws := &WritableStream{
		destroyed:     false,
		ended:         0,
		highWaterMark: conf.highWaterMark,
		buffer:        NewBuffer(),
		doneCh:        make(chan struct{}),
		writing:       0,
		draining:      0,
	}

	out := conf.output

	if out == nil {
		return nil, errors.New("writable stream config requires output type")
	}

	ws.On("write", func(evt types.Event) {
		// Call output.Write with data
		err := out.Write(evt.Data)

		if err != nil {
			log.Fatal(err.Error())
		}

		afterWrite(ws)
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
		ws.destroyed = true

		// unset if draining
		if isDraining(ws) {
			atomic.CompareAndSwapUint32(&ws.draining, 1, 0)
		}

		// Send message to done channel
		close(ws.doneCh)
	})

	return ws, nil
}

func canWriteMore(ws *WritableStream) bool {
	return !writableEnded(ws) &&
		uint(ws.buffer.Len()) < ws.highWaterMark &&
		!isDraining(ws)
}

func writableEnded(ws *WritableStream) bool {
	return atomic.LoadInt32(&ws.ended) == 1
}

func writableBufferChunk(ws *WritableStream, chunk types.Chunk) {
	ws.buffer.Write(chunk)

	if uint(ws.buffer.Len()) >= ws.highWaterMark && !isDraining(ws) {
		atomic.CompareAndSwapUint32(&ws.draining, 0, 1)
	}
}

func isDraining(ws *WritableStream) bool {
	return atomic.LoadUint32(&ws.draining) == 1
}

func isWriting(ws *WritableStream) bool {
	return atomic.LoadUint32(&ws.writing) == uint32(1)
}

func acquireWrite(ws *WritableStream) bool {
	return atomic.CompareAndSwapUint32(&ws.writing, 0, 1)
}

func releaseWrite(ws *WritableStream) bool {
	return atomic.CompareAndSwapUint32(&ws.writing, 1, 0)
}

func endWritable(ws *WritableStream) bool {
	if atomic.CompareAndSwapInt32(&ws.ended, 0, 1) {
		return true
	}

	return false
}

func writableDestroyed(ws *WritableStream) bool {
	return ws.destroyed == true
}

func writableFinished(ws *WritableStream) bool {
	ws.mux.RLock()
	defer ws.mux.RUnlock()
	return writableEnded(ws) && ws.buffer.Len() == 0
}

func afterWrite(ws *WritableStream) {
	// Emit drain event if needed
	if needDrain(ws) {
		if atomic.CompareAndSwapUint32(&ws.draining, 1, 0) {
			go ws.Emit("drain", nil)
		}
	}
	if writableFinished(ws) {
		ws.Emit("writable_done", nil)
	}
}

func needDrain(ws *WritableStream) bool {
	buffLen := ws.buffer.Len()
	return !writableEnded(ws) && buffLen == 0 && isDraining(ws)
}
