package stream

import "sync"

import "github.com/jpg013/go_stream/types"

type TransformState struct {
	transforming  int32
	needTransform bool
	writeChunk    types.Chunk
	writeCb       func(error)
	mutex         sync.RWMutex
}

func NewTransformState() *TransformState {
	return &TransformState{
		transforming:  0,
		needTransform: false,
	}
}
