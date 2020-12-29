package main

import (
	"github.com/jpg013/go_stream/types"
)

type InternalState struct {
	// buffer that holds data chunks that has been either
	// read from a Readable or waiting to be written
	// to a Writable
	buffer chan types.Chunk

	// represents the stream mode
	mode uint32

	highWaterMark uint

	doneCh chan (struct{})

	// flag determines whether the writable stream is ended
	writableEnded int32

	// flag determines whether the readable stream is ended
	readableEnded int32

	// flag determines whether the writable stream is destroyed
	writableDestroyed int32

	// flag determines whether the readble stream is destroyed
	readableDestroyed int32

	// flag determines whether the stream is currently reading
	reading uint32

	// flag determines whether the stream is currently writing
	writing uint32

	// determines whether the stream is currently draining
	draining uint32

	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest Writable

	awaitDrainWriters uint32
}

func NewInternalState(conf StreamConfig) *InternalState {
	return &InternalState{
		buffer:            make(chan types.Chunk, conf.highWaterMark),
		mode:              ReadableNull,
		highWaterMark:     conf.highWaterMark,
		doneCh:            make(chan struct{}),
		writableEnded:     0,
		readableEnded:     0,
		writableDestroyed: 0,
		readableDestroyed: 0,
		reading:           0,
		writing:           0,
		draining:          0,
		dest:              nil,
		awaitDrainWriters: 0,
	}
}
