package stream

import (
	"container/list"
	"sync"
)

type WritableState struct {
	buffer        *list.List
	highWaterMark int
	length        int32
	draining      uint32
	ended         bool
	destroyed     bool
	writing       uint32
	mux           sync.RWMutex
}

func NewWritableState() *WritableState {
	return &WritableState{
		buffer:        list.New(),
		length:        0,
		highWaterMark: 3,
		writing:       0,
		draining:      0,
		ended:         false,
		destroyed:     false,
	}
}
