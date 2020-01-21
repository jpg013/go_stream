package stream

import "github.com/jpg013/go_stream/types"

type TransformState struct {
	transforming  uint32
	writeChunk    types.Chunk
	writeCb       func(error)
	needTransform bool
}

func NewTransformState() *TransformState {
	return &TransformState{
		writeChunk:    nil,
		writeCb:       nil,
		transforming:  0,
		needTransform: false,
	}
}
