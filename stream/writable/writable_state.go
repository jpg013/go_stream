package writable

import (
	"container/list"
	"sync"
)

type WritableState struct {
	buffer         *list.List
	highWaterMark  int
	draining       bool
	writeRequested bool
	ended          bool
	destroyed      bool
	mtx            sync.Mutex
}

func NewWritableState() *WritableState {
	return &WritableState{
		buffer:         list.New(),
		highWaterMark:  3,
		draining:       false,
		ended:          false,
		writeRequested: false,
		destroyed:      false,
	}
}
