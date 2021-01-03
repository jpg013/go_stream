package main

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

type Transform interface {
	Stream
	Read() types.Chunk
	Write(types.Chunk) bool
}

type TransformStream struct {
	// embedded event emitter
	emitter.Emitter
	state *InternalState
}

func (ts *TransformStream) Pipe(w Writable) Writable {
	return w
}

func (ts *TransformStream) Write(data types.Chunk) bool {
	// if ts.isEnded() {
	// 	log.Fatal("Cannot write to ended stream")
	// }

	// if ts.isDestroyed() {
	// 	log.Fatal("Cannot write to destroyed stream")
	// }

	// // Similar to a readable stream, a nil chunk indicates end of stream.
	// if data == nil {
	// 	writableEndOfChunk(ws)
	// 	return false
	// } else {
	// 	writableBufferChunk(ws, data)
	// }

	// doWrite := !isWriting(ws) && !writableEnded(ws)

	// if doWrite {
	// 	write(ws)
	// }

	return true
}

func (ts *TransformStream) Read() (data types.Chunk) {
	return data
}

func (ts *TransformStream) Done() <-chan struct{} {
	return ts.state.doneCh
}

func NewTransformStream(conf StreamConfig) Transform {
	ts := &TransformStream{}
	return ts
}
