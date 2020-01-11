package transform

import (
	"container/list"
	"sync"
)

type TransformState struct {
	buffer         *list.List
	highWaterMark  int
	draining       bool
	writeRequested bool
	ended          bool
	destroyed      bool
	mtx            sync.Mutex
}

func NewTransformState() *TransformState {
	return &TransformState{
		buffer:        list.New(),
		highWaterMark: 3,
	}
}
