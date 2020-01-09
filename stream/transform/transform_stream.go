package transform

import operator "github.com/jpg013/go_stream/operators"

import "github.com/jpg013/go_stream/types"

func (ts *Stream) Pipe(w types.Writable) types.Writable {
	return nil
}

func (ts *Stream) Read() bool {
	return false
}

func (ts *Stream) Write(types.Chunk) bool {
	return false
}

func (ts *Stream) Done() <-chan struct{} {
	return ts.doneChan
}

func NewTransform(op operator.Type) (types.Transform, error) {
	return nil, nil
}
