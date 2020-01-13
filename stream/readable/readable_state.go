package readable

import "sync"

import "container/list"

type ReadableState struct {
	buffer        *list.List
	mode          ReadableMode
	highWaterMark int
	ended         bool
	destroyed     bool
	readRequested bool
	mux           sync.Mutex
}

func NewReadableState() *ReadableState {
	return &ReadableState{
		buffer:        list.New(),
		mode:          ReadableNull,
		destroyed:     false,
		ended:         false,
		highWaterMark: 5,
		readRequested: false,
	}
}
