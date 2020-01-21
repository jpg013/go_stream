package stream

// A transform stream is a special kind of stream that is both a readable and
// a writable stream.

import (
	"fmt"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func (ts *TransformStream) Pipe(w Writable) Writable {
	return attachWritable(ts, w)
}

func (ts *TransformStream) Done() <-chan struct{} {
	return ts.doneChan
}

func (stream *TransformStream) Read() types.Chunk {
	rs := stream.GetReadableState()

	// if stream is destroyed then there is no point.
	if readableDestroyed(stream) {
		return nil
	}

	// If we've ended, and we're now clear, then exit.
	if readableFinished(stream) {
		// All pending data has flushed to the writer
		endReadable(stream)
		return nil
	}

	len := int(atomic.LoadInt32(&rs.length))

	// Always try to read more if buffer is not full
	doRead := len == 0 || len < rs.highWaterMark

	// However, if we are already ready, or ended or destroyed then don't read more.
	if isReading(stream) || readableEnded(stream) || readableDestroyed(stream) {
		doRead = false
	}

	if doRead {
		stream.read()
	}

	var data types.Chunk

	if !awaitingDrain(stream) && !isTransforming(stream) {
		data = fromReadableBuffer(rs)

		if data != nil {
			emitData(stream, data)
		}
	}

	return data
}

func isTransforming(t Transform) bool {
	ts := t.GetTransformState()
	return atomic.LoadUint32(&ts.transforming) == 1
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

	stream := &TransformStream{
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
					afterTransform(stream, resp)
				}
			case err := <-errorChan:
				panic(err)
			}
		}
	}()

	stream.read = func() {
		ts := stream.state

		ts.mutex.RLock()
		chunk := ts.writeChunk
		cb := ts.writeCb
		ts.mutex.RUnlock()

		// If we can transform then go, else set needTransform flag and continue
		if chunk != nil && cb != nil && acquireTransform(stream) {
			transform(stream)
		} else {
			ts.needTransform = true
		}
	}

	stream.On("data", func(evt types.Event) {
		rs := stream.GetReadableState()
		ret := rs.dest.Write(evt.Data)

		if !ret {
			// it is possible to get in a permanent state of "paused", when the
			// dest has unpiped from the source.
			pauseReadable(stream)

			if atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 0, 1) {
				rs.dest.Once("drain", func(evt types.Event) {
					atomic.CompareAndSwapUint32(&rs.awaitDrainWriters, 1, 0)
					resumeReadable(stream)
				})
			}
		}

		// After each data event, decrement the pendingReads count by 1.
		if atomic.AddInt32(&rs.pendingReads, -1) < 0 {
			panic("Cannot have negative pending reads")
		}
	})

	stream.Once("writable_end", func(evt types.Event) {
		readableAddChunk(stream, nil)
	})

	stream.Once("readable_end", func(evt types.Event) {
		rs := stream.readableState
		rs.dest.Write(nil)
		rs.destroyed = true
		close(stream.doneChan)
	})

	stream.On("write", func(evt types.Event) {
		ts := stream.state
		rs := stream.GetWriteState()

		if atomic.LoadUint32(&rs.writing) != 1 {
			panic("We can't be in the write like this??")
		}

		if atomic.LoadUint32(&ts.transforming) == 1 {
			fmt.Println("this is really bad :(")
			fmt.Println(evt.Data)
			fmt.Println(ts.writeChunk)
			panic("What the fork!")
		}

		ts.mutex.Lock()
		ts.writeChunk = evt.Data
		ts.writeCb = func(err error) {
			if err != nil {
				panic(err)
			}

			afterWrite(stream)
		}
		ts.mutex.Unlock()

		// TODO: Only read if buffer is not full or needTransform flag set
		stream.read()
	})

	return stream, nil
}

func afterTransform(t Transform, data types.Chunk) {
	ts := t.GetTransformState()
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

func transform(stream *TransformStream) {
	ts := stream.state
	ts.mutex.RLock()
	cb := ts.writeCb
	chunk := ts.writeChunk
	ts.mutex.RUnlock()

	if cb == nil {
		panic("nil transform callback")
	}

	if chunk == nil {
		panic("nil transform chunk")
	}

	if !isTransforming(stream) {
		panic("transform flag not set")
	}

	// Execute transform operator with write chunk
	err := stream.operator.Exec(chunk)

	ts.mutex.Lock()
	ts.writeChunk = nil
	ts.writeCb = nil
	ts.mutex.Unlock()

	if !releaseTransform(stream) {
		panic("cannot release transform")
	}

	// call callback with error
	cb(err)
}
