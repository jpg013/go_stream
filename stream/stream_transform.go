package stream

// A transform stream is a special kind of stream that is both a readable and
// a writable stream.

import (
	"fmt"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func (ts *TransformStream) Pipe(Writable) Writable {
	return nil
}

func (ts *TransformStream) Done() <-chan struct{} {
	return ts.doneChan
}

func (ts *TransformStream) Read() types.Chunk {
	return nil
}

func isTransforming(ts *TransformStream) bool {
	return atomic.LoadUint32(&ts.state.transforming) == 1
}

func (ts *TransformStream) Write(chunk types.Chunk) bool {
	state := ts.writableState

	if state.destroyed {
		panic("Error stream destroyed")
	}

	if chunk == nil {
		writableEndOfChunk(ts)
		return false
	}

	writeOrBuffer(ts, chunk)

	return canWriteMore(ts)
}

func (ts *TransformStream) GetTransformState() *TransformState {
	return ts.state
}

func (ts *TransformStream) GetWriteState() *WritableState {
	return ts.writableState
}

func NewTransformStream(config *Config) (Transform, error) {
	doneChan := make(chan struct{})

	op := config.Transform.Operator

	if op == nil {
		panic("Transform stream requires non-nil operator")
	}

	ts := &TransformStream{
		doneChan:      doneChan,
		state:         NewTransformState(),
		writableState: NewWritableState(),
	}

	go func() {
		outChan := op.GetOutput()
		errorChan := op.GetError()

		for {
			select {
			case resp := <-outChan:
				fmt.Println("We have a chunk", resp)
			case err := <-errorChan:
				panic(err)
			}
		}
	}()

	ts.transform = func(chunk types.Chunk) {
		if !atomic.CompareAndSwapUint32(&ts.state.transforming, 0, 1) {
			panic("bad bad")
		}

		// Execute chunk
		op.Exec(chunk)

		atomic.CompareAndSwapUint32(&ts.state.transforming, 0, 1)
	}

	return ts, nil
}
