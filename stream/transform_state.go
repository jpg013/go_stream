package stream

import "github.com/jpg013/go_stream/types"

import "sync"

type TransformState struct {
	transforming  uint32
	writeChunk    types.Chunk
	writeCb       func(error)
	needTransform bool
	mutex         sync.RWMutex
}

func NewTransformState() *TransformState {
	return &TransformState{
		writeChunk:    nil,
		writeCb:       nil,
		transforming:  0,
		needTransform: false,
	}
}
