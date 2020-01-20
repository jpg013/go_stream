package writable

import (
	"container/list"
)

type WritableState struct {
	buffer        *list.List
	highWaterMark int
	length        int32
	draining      bool
	ended         bool
	destroyed     bool
	writing       bool
}

func NewWritableState() *WritableState {
	return &WritableState{
		buffer:        list.New(),
		length:        0,
		highWaterMark: 3,
		writing:       false,
		draining:      false,
		ended:         false,
		destroyed:     false,
	}
}
