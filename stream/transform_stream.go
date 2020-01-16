package stream

// import (
// 	"fmt"

// 	"github.com/jpg013/go_stream/emitter"
// 	"github.com/jpg013/go_stream/types"
// )

// func NewTransformStream(op types.Operator) (types.Transform, error) {
// 	ts := &Stream{
// 		writableState:  NewWritableState(),
// 		transformState: NewTransformState(),
// 		emitter:        emitter.NewEmitter(),
// 		Type:           types.TransformType,
// 		doneChan:       make(chan struct{}),
// 	}

// 	// go func() {
// 	// 	for {
// 	// 		select {
// 	// 		case chunk, ok := <-op.GetOutput():
// 	// 			if !ok {
// 	// 				return
// 	// 			}
// 	// 			fmt.Println("WE have a chunk", chunk, ok)
// 	// 		case err, ok := <-op.GetError():
// 	// 			if ok {
// 	// 				panic(err.Error())
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }()

// 	// ts.On("end", func(evt types.Event) {
// 	// 	state := ts.writableState
// 	// 	state.mux.Lock()
// 	// 	defer state.mux.Unlock()

// 	// 	if state.destroyed {
// 	// 		panic("cannot end writable stream, already destroyed")
// 	// 	}

// 	// 	if !state.ended {
// 	// 		panic("\"end\" event emitted before end of stream")
// 	// 	}

// 	// 	// Call end on the operator
// 	// 	op.End()

// 	// 	// Set destroyed to be true
// 	// 	state.destroyed = true

// 	// 	// // set draining to false if draining
// 	// 	if state.draining {
// 	// 		state.draining = false
// 	// 	}

// 	// 	close(ts.doneChan)
// 	// })

// 	ts.On("write", func(evt types.Event) {
// 		// state := ts.writableState
// 		// state.mux.Lock()
// 		// defer state.mux.Unlock()

// 		// if !state.writing {
// 		// 	panic("should not be in write event handler without write requested")
// 		// }

// 		// // call Transform on the functoin
// 		// ts.Transform()

// 		// // Call operator.Exec with data chunk
// 		// op.Exec(evt.Data)

// 		// // unset the write requested
// 		// state.writeRequested = false

// 		// // Check if stream is ended
// 		// checkWritableEndOfStream(ts)

// 		// // Emit drain event if needed
// 		// if state.buffer.Len() == 0 && state.draining && !state.ended {
// 		// 	ts.emitter.Emit("drain", nil)
// 		// 	state.draining = false
// 		// }

// 		// maybeWriteMore(ts)
// 	})

// 	return ts, nil
// }

// func (ts *Stream) write(chunk types.Chunk) {
// 	// state := ts.transformState
// 	// state.mux.Lock()
// 	// defer state.mux.Unlock()

// 	if !ts.writableState.writing {
// 		panic("should not be in write event handler without write requested or valid data chunk")
// 	}

// 	if ts.transformState.writeChunk != nil {
// 		// What to do in this case where there is a write chunk pending writing??
// 		panic("transform write Chunk should not be null")
// 	}

// 	// set the write chunk and call read if we want to read data
// 	ts.transformState.writeChunk = chunk
// 	ts.read()
// }

// // interal read function, responsible for providing data to the
// // underlying writable stream.
// func (ts *Stream) read() {
// 	chunk := ts.transformState.writeChunk

// 	if chunk != nil && !ts.transformState.transforming {
// 		// Do transform here.
// 		ts.transform(chunk)
// 	} else {
// 		ts.transformState.needTransform = true
// 	}
// }

// func (ts *Stream) transform(chunk types.Chunk) {
// 	// Do some transform shit here
// 	fmt.Println("IN the transformer", chunk.(int))

// 	afterTransform(ts, chunk)
// }

// func (ts *Stream) Pipe(w types.Writable) types.Writable {
// 	// IMPLEMENT
// 	return nil
// }

// func (ts *Stream) Read() bool {
// 	return false
// }

// // On function
// func (ts *Stream) On(topic string, fn types.EventHandler) {
// 	ts.emitter.On(topic, fn)
// }

// // Emit function
// func (ts *Stream) Emit(topic string, data interface{}) {
// 	ts.emitter.Emit(topic, data)
// }

// func (ts *Stream) Done() <-chan struct{} {
// 	return ts.doneChan
// }

// func writableEndOfChunk(ts *Stream) {
// 	writeState := ts.writableState

// 	if writeState.ended {
// 		panic("stream already ended")
// 	}

// 	// Set the ended flag to be true
// 	writeState.ended = true

// 	// Call maybeWriteMore() in case we need to continue
// 	// writing until buffer is drained
// 	maybeWriteMore(ts)
// }

// func afterTransform(ts *Stream, chunk types.Chunk) {
// 	state := ts.transformState

// 	state.transforming = false
// 	state.writeChunk = nil

// }

// // func afterWrite(ts *Stream) {
// // 	writeState := ts.writableState

// // 	// Check to see if the write stream has ended or we are writing or if there is
// // 	// more "buffered" data needing to be written. If not, then
// // 	// emit the "end" event
// // 	if writeState.ended &&
// // 		writeState.buffer.Len() == 0 &&
// // 		!writeState.writing {
// // 		ts.Emit("end", nil)
// // 	}

// // 	maybeWriteMore(ts)
// // }

// func (ts *Stream) Write(data types.Chunk) bool {
// 	// Lock the write state, requires writing to be synchronized
// 	writeState := ts.writableState
// 	writeState.mux.Lock()
// 	defer writeState.mux.Unlock()

// 	// If writing to destroyed stream then Panic
// 	if writeState.destroyed {
// 		panic("error writing to stream, stream is destroyed")
// 	}

// 	// nil chunk indicates end of stream. Call onEndOfChunk and return
// 	// false to indicate no more writing to take place
// 	if data == nil {
// 		writableEndOfChunk(ts)
// 	} else {
// 		writeOrBuffer(ts, data)
// 	}

// 	return canWriteMore(ts)
// }

// // If we're already writing something, then just put this
// // in the queue, and wait our turn. Otherwise, call doWrite
// // If we return false, then we need a drain event, so set that flag.
// func writeOrBuffer(ts *Stream, chunk types.Chunk) {
// 	state := ts.writableState

// 	// If there is not data buffered, and no write requested,
// 	// then we can simply call emitWritable, and continue.
// 	if !state.writing && state.buffer.Len() == 0 {
// 		emitWritable(ts, chunk)
// 	} else {
// 		writableBufferChunk(ts, chunk)
// 	}
// }

// func writableBufferChunk(ts *Stream, data types.Chunk) {
// 	// state := ts.writableState
// 	// state.buffer.PushBack(data)

// 	// if state.buffer.Len() >= state.highWaterMark && !state.draining {
// 	// 	state.draining = true
// 	// }
// }

// func emitWritable(ts *Stream, data types.Chunk) {
// 	state := ts.writableState

// 	if state.writing {
// 		panic("cannot call write on WritableStream, writing in progress")
// 	}

// 	if data == nil {
// 		panic("cannot write nil data chunk")
// 	}

// 	// Set writing to true
// 	state.writing = true

// 	// Emit the write event with data chunk
// 	ts.Emit("write", data)
// }

// // Conditions for writing more are the following:
// // if there is not a current write chunk being written and
// // if the write buffer has data in it and the stream has not been
// // destroyed.
// func maybeWriteMore(ts *Stream) {
// 	writeState := ts.writableState

// 	if !writeState.writing || writeState.buffer.Len() == 0 || writeState.destroyed {
// 		return
// 	}

// 	chunk := fromBuffer(ts.writableState.buffer)

// 	emitWritable(ts, chunk)
// }

// // Function to determine whether the writable stream can receive more
// // data from the readable source. This should be called to provide a return
// // value from the Write(Chunk) method to signal to the readable stream whether
// // the writable is accepting any more data.
// func canWriteMore(ts *Stream) bool {
// 	state := ts.writableState

// 	// If were ended or destroyed obviously cannot receive more data
// 	if state.ended ||
// 		state.destroyed ||
// 		state.buffer.Len() >= state.highWaterMark {
// 		return false
// 	}

// 	return true
// }
