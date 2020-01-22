package stream

// A transform stream is a special kind of stream that is both a readable and
// a writable stream.

import (
	"fmt"
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func (stream *TransformStream) Pipe(w Writable) Writable {
	attachWritable(stream, w)

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

	return w
}

func (ts *TransformStream) Done() <-chan struct{} {
	return ts.doneChan
}

func (stream *TransformStream) Read() types.Chunk {
	// if stream is destroyed then there is no point.
	if readableDestroyed(stream) {
		return nil
	}

	// ts := stream.GetTransformState()
	rs := stream.GetReadableState()
	len := int(atomic.LoadInt32(&rs.length))

	// If we've ended, and we're now clear, then exit.
	if readableFinished(stream) {
		endReadable(stream)
		return nil
	}

	// // Always try to read more if buffer is not full
	doRead := len == 0 || len < rs.highWaterMark

	// However, if we are already ready, or ended or destroyed then don't read more.
	if isReading(stream) ||
		readableEnded(stream) ||
		readableDestroyed(stream) {
		doRead = false
	}

	if doRead {
		stream.read()
	}

	var data types.Chunk

	if !awaitingDrain(stream) {
		data = fromReadableBuffer(rs)

		if data != nil && rs.dest != nil {
			emitData(stream, data)
		}
	}

	return data
}

func isTransforming(t Transform) bool {
	ts := t.GetTransformState()
	return atomic.LoadInt32(&ts.transforming) > 0
}

func (ts *TransformStream) Write(chunk types.Chunk) bool {
	ws := ts.writableState

	if ws.destroyed {
		panic("Error stream destroyed")
	}

	if chunk == nil {
		writableEndOfChunk(ts)
		len := atomic.LoadInt32(&ws.length)
		writing := atomic.LoadUint32(&ws.writing) > 0
		needEnd := writableEnded(ts) && len == 0 && !writing

		if needEnd {
			endWritable(ts)
		}

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

	op := config.Transform.Mapper

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

	stream.read = func() {
		if !acquireRead(stream) {
			fmt.Println("transform read error, unable to acquire stream read.")
			return
		}
		ts := stream.state
		if ts.writeChunk != nil && ts.writeCb != nil {
			// Transform the write chunk and continue
			transform(stream)
		} else {
			if !releaseRead(stream) {
				return
			}
			ts.needTransform = true
			// putting this line here because sometimes we deadlock without it.
			// go maybeReadMore(stream)
		}
	}

	stream.Once("writable_end", func(evt types.Event) {
		// Tell the transformer we are done writing
		// if isTransforming(stream) {
		// 	fmt.Println("writable_end is waiting for transform")
		// 	time.Sleep(1 * time.Second)
		// }
		// readableAddChunk(stream, nil)
		// stream.operator.End()
	})

	stream.Once("readable_end", func(evt types.Event) {
		rs := stream.readableState

		if rs.dest != nil {
			rs.dest.Write(nil)
		}

		rs.destroyed = true
		close(stream.doneChan)
	})

	stream.On("write", func(evt types.Event) {
		ts := stream.state
		rs := stream.GetWriteState()

		if atomic.LoadUint32(&rs.writing) != 1 {
			panic("We can't be in the write like this??")
		}

		var called uint32

		if ts.writeChunk != nil || ts.writeCb != nil {
			panic("WJSKDFSJKFLSDLKFJKSLDFLKJ")
		}

		cb := func(err error) {
			if !atomic.CompareAndSwapUint32(&called, 0, 1) {
				panic("CB CALLED TWICE")
			}

			// sanity check to remove later
			if ts.writeChunk == nil || ts.writeCb == nil {
				panic("FAIL SANITY")
			}

			if ts.writeChunk != evt.Data {
				panic("FAIL WRITE CHUNK SANITY")
			}

			if err != nil {
				panic(err)
			}

			ts.writeChunk = nil
			ts.writeCb = nil

			afterWrite(stream)
		}

		ts.writeChunk = evt.Data
		ts.writeCb = cb

		// Inc transforming count
		atomic.AddInt32(&ts.transforming, 1)

		// TODO: Only read if buffer is not full or needTransform flag set
		stream.read()
	})

	return stream, nil
}

func afterTransform(t Transform, data types.Chunk, err error) {
	ts := t.GetTransformState()
	cb := ts.writeCb

	if cb == nil {
		panic("BAD CALLBACK")
	}

	atomic.AddInt32(&ts.transforming, -1)
	// fire callback with error
	cb(err)

	// unset us from reading...
	if !releaseRead(t) {
		panic("Could not unset reading flag")
	}
	readableAddChunk(t, data)
	ts.needTransform = false

	if writableFinished(t) {
		// If the writable side of the transform stream is done push
		// a nil chunk to the readable stream to end
		endReadable(t)
		go t.Read()
	}
}

func transform(stream *TransformStream) {
	ts := stream.GetTransformState()
	chunk := ts.writeChunk

	if chunk == nil {
		panic("OH MY GOD")
	}

	// Execute transform operator with write chunk
	transformedData, err := stream.operator(chunk)

	afterTransform(stream, transformedData, err)
}

// func acquireTransform(stream Transform) bool {
// 	ts := stream.GetTransformState()
// 	return atomic.CompareAndSwapUint32(&ts.transforming, 0, 1)
// }

// func releaseTransform(stream Transform) bool {
// 	ts := stream.GetTransformState()
// 	return atomic.CompareAndSwapUint32(&ts.transforming, 1, 0)
// }
