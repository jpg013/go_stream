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

func (ts *TransformStream) GetReadableState() *ReadableState {
	return ts.readableState
}

func NewTransformStream(config *Config) (Transform, error) {
	doneChan := make(chan struct{})

	op := config.Transform.Operator

	if op == nil {
		panic("Transform stream config requires valid operator")
	}

	ts := &TransformStream{
		doneChan:      doneChan,
		state:         NewTransformState(),
		writableState: NewWritableState(),
		readableState: NewReadableState(),
		operator:      op,
	}

	go func() {
		outChan := op.GetOutput()
		errorChan := op.GetError()

		for {
			select {
			case resp, ok := <-outChan:
				if ok {
					afterTransform(ts, resp)
				}
			case err := <-errorChan:
				panic(err)
			}
		}
	}()

	ts.read = func() {
		state := ts.state

		// If we can transform then go, else set needTransform flag and continue
		if state.writeChunk != nil && acquireTransform(ts) {
			transform(ts)
		} else {
			state.needTransform = true
		}
	}

	ts.On("write", func(evt types.Event) {
		state := ts.state

		if atomic.LoadUint32(&state.transforming) == 1 {
			panic("What the fork!")
		}

		state.writeChunk = evt.Data
		state.writeCb = func(err error) {
			if err != nil {
				panic(err)
			}

			afterWrite(ts)
		}

		// TODO: Only read if buffer is not full or needTransform flag set
		ts.read()
	})

	return ts, nil
}

func afterTransform(t Transform, data types.Chunk) {
	ts := t.GetTransformState()
	fmt.Println("after transform: ", data)
	readableAddChunk(t, data)
	ts.needTransform = false

}

func acquireTransform(t Transform) bool {
	state := t.GetTransformState()

	return atomic.CompareAndSwapUint32(&state.transforming, 0, 1)
}

func releaseTransform(t Transform) bool {
	state := t.GetTransformState()

	return atomic.CompareAndSwapUint32(&state.transforming, 1, 0)
}

func transform(ts *TransformStream) {
	state := ts.state

	cb := state.writeCb

	if cb == nil {
		panic("nil transform callback")
	}

	chunk := state.writeChunk

	if chunk == nil {
		panic("nil transform chunk")
	}

	if atomic.LoadUint32(&state.transforming) != 1 {
		panic("transform flag not set")
	}

	// Execute transform operator with write chunk
	err := ts.operator.Exec(chunk)

	if !releaseTransform(ts) {
		panic("cannot release transform")
	}

	// call callback with error
	cb(err)
}
