package stream

import "container/list"

import "sync"

type ReadableState struct {
	mux               sync.RWMutex
	buffer            *list.List
	mode              uint32
	pendingReads      int32
	length            int32
	reading           uint32
	highWaterMark     int
	ended             bool
	destroyed         bool
	awaitDrainWriters uint32
	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest Writable
}

func NewReadableState() *ReadableState {
	return &ReadableState{
		buffer:            list.New(),
		mode:              ReadableNull,
		length:            0,
		pendingReads:      0,
		highWaterMark:     16,
		destroyed:         false,
		reading:           0,
		ended:             false,
		awaitDrainWriters: 0,
		dest:              nil,
	}
}
