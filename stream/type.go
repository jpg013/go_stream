package stream

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

type ReadableMode uint32

const (
	ReadableType StreamType = iota
	TransformType
	WritableType
)

type StreamType uint32

const (
	// ReadableFlowing mode reads from the underlying system automatically
	// and provides to an application as quickly as possible.
	ReadableFlowing uint32 = 1
	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
	ReadableNotFlowing uint32 = 2
	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
	ReadableNull uint32 = 0
)

type WritableStream struct {
	emitter.Emitter
	state    *WritableState
	Type     StreamType
	doneChan chan struct{}
}

// Stream represents a readable stream type
type ReadableStream struct {
	// embedded event emitter
	emitter.Emitter
	state *ReadableState
	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest       Writable
	doneChan   chan struct{}
	StreamType StreamType
	// internal read method that can be overwritten
	read func()
}

type TransformStream struct {
	// embedded event emitter
	emitter.Emitter
	doneChan      chan struct{}
	state         *TransformState
	writableState *WritableState
	transform     func(types.Chunk)
}

type Stream interface {
	types.EventEmitter
	Pipe(Writable) Writable
	Done() <-chan struct{}
}

type Readable interface {
	Stream
	Read() types.Chunk
	GetReadState() *ReadableState
}

type Writable interface {
	Stream
	Write(types.Chunk) bool
	GetWriteState() *WritableState
}

type Transform interface {
	Stream
	Read() types.Chunk
	Write(types.Chunk) bool
	GetTransformState() *TransformState
}
