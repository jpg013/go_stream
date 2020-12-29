package main

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/generators"
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

	buffer        chan types.Chunk
	mode          uint32
	highWaterMark uint
	doneCh        chan (struct{})
	gen           generators.Type

	Type StreamType

	// writable ended bit
	wEnded int32

	// destroyed bit
	wDestroyed int32

	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest Writable

	// bit determining whether the stream is currently reading or not
	reading uint32

	// internal read method that can be overwritten
	read func()
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
	return ts.doneCh
}

func NewTransformStream() Transform {
	ts := &TransformStream{}
	return ts
}
