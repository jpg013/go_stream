package readable

import "container/list"

type ReadableState struct {
	buffer        *list.List
	mode          int32
	highWaterMark int
	pendingData   int32
	ended         bool
	destroyed     bool
	reading       bool
}

func NewReadableState() *ReadableState {
	return &ReadableState{
		buffer:        list.New(),
		mode:          ReadableNull,
		pendingData:   0,
		highWaterMark: 5,
		destroyed:     false,
		reading:       false,
		ended:         false,
	}
}
