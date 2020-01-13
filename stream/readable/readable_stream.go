package readable

import (
	"container/list"
	"sync"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

func maybeReadMore(rs *Stream) {
	state := rs.state

	if state.destroyed {
		return
	}

	// If we are ended, and there is no more data to be written to
	// output, then exit.
	if state.ended && state.buffer.Len() == 0 {
		return
	}

	// We keep reading while we are flowing or until buffer is full
	if isReadableFlowing(rs) || state.buffer.Len() < state.highWaterMark {
		go rs.Read()
	}
}

func isReadableFlowing(rs *Stream) bool {
	return rs.state.mode == ReadableFlowing
}

// Manually shove something into the read() buffer.
// This returns true if the highWaterMark has not yet been hit,
// similar to how writable.Write() returns true if you should write some more.
func readableAddChunk(rs *Stream, data types.Chunk) {
	state := rs.state

	// read chunk is nil, signals end of stream.
	if data == nil {
		onEndOfChunk(rs)
	} else {
		if isReadableFlowing(rs) &&
			!state.readRequested &&
			state.buffer.Len() == 0 {
			emitReadable(rs, data)
		} else {
			state.buffer.PushBack(data)
		}
	}

	maybeReadMore(rs)
}

func pauseReadable(rs *Stream) {
	if rs.state.mode == ReadableFlowing {
		rs.state.mode = ReadableNotFlowing
	}
}

func resumeReadable(rs *Stream) {
	state := rs.state

	// if state is destroyed then no point in continuing
	if state.destroyed {
		return
	}

	if rs.state.mode != ReadableFlowing {
		rs.state.mode = ReadableFlowing
		go rs.Read()
	}
}

func fromBuffer(buf *list.List) interface{} {
	val := buf.Front()

	if val == nil {
		return val
	}

	buf.Remove(val)

	return val.Value
}

// Set the ended flag
func onEndOfChunk(rs *Stream) {
	if rs.state.ended {
		panic("stream is already ended")
	}
	rs.state.ended = true
	checkEndOfStream(rs)
}

type Stream struct {
	emitter  *emitter.Type
	state    *ReadableState
	dest     types.Writable
	mux      sync.Mutex
	gen      types.Generator
	doneChan chan struct{}
	Type     types.StreamType
}

func canReadMore(rs *Stream) bool {
	state := rs.state

	return !state.destroyed &&
		isReadableFlowing(rs) &&
		state.buffer.Len() < state.highWaterMark
}

func checkEndOfStream(rs *Stream) {
	state := rs.state
	// Check to see if the readable stream has ended.
	if state.ended && state.buffer.Len() == 0 && !state.readRequested {
		rs.Emit("end", nil)
	}
}

func emitReadable(rs *Stream, data types.Chunk) {
	state := rs.state

	// if a read has been requested then exit
	if state.readRequested || !isReadableFlowing(rs) || data == nil {
		panic("cannot emit readable")
	}

	// Set read requested to true and emit the data
	state.readRequested = true
	// emit chunk and continue
	rs.Emit("data", data)
}

func (rs *Stream) Read() bool {
	state := rs.state
	state.mux.Lock()
	defer state.mux.Unlock()

	// if stream is destroyed then there is not point, simply exit
	if state.destroyed {
		return false
	}

	// If we have data in the buffer and we are flowing and a read has not been requested
	// then call emitReadable to write data to the underlying writable stream
	if isReadableFlowing(rs) &&
		state.buffer.Len() > 0 &&
		!state.readRequested {
		emitReadable(rs, fromBuffer(rs.state.buffer))
	}

	// If the readable stream has not ended and can push more data into the buffer the continue
	if !state.ended && state.buffer.Len() < state.highWaterMark {
		chunk, _ := rs.gen.Next()
		readableAddChunk(rs, chunk)
	}

	return canReadMore(rs)
}

func (rs *Stream) On(topic string, fn types.EventHandler) {
	rs.emitter.On(topic, fn)
}

func (rs *Stream) Emit(topic string, data interface{}) {
	rs.emitter.Emit(topic, data)
}

func (rs *Stream) Pipe(w types.Writable) types.Writable {
	if rs.state.ended {
		panic("cannot Pipe() to ended readable")
	}

	if rs.state.destroyed {
		panic("cannot Pipe() to destroyed readable")
	}

	if rs.dest != nil {
		panic("Cannot pipe Writable - Readable destination already exists")
	}

	// Assign writable to source destination
	rs.dest = w

	state := rs.state

	rs.On("data", func(evt types.Event) {
		state.mux.Lock()
		defer state.mux.Unlock()

		if !isReadableFlowing(rs) {
			panic("readable stream is not flowing")
		}

		if !rs.dest.Write(evt.Data) {
			pauseReadable(rs)
		}

		// unset the read requested flag after write
		if rs.state.readRequested {
			rs.state.readRequested = false
		}

		// Check the end of the stream, if we are ended
		// and there is not more data to emit we can emit the
		// "end" event
		checkEndOfStream(rs)

		// Trying reading more if we can
		maybeReadMore(rs)
	})

	rs.On("end", func(evt types.Event) {
		if rs.state.destroyed {
			panic("we have already been here")
		}

		// Write a nil chunk to the destination to signal end
		rs.dest.Write(nil)
		rs.state.destroyed = true
		close(rs.doneChan)
	})

	w.On("drain", func(evt types.Event) {
		resumeReadable(rs)
	})

	// Resume reading
	resumeReadable(rs)

	return w
}

func (rs *Stream) Done() <-chan struct{} {
	return rs.doneChan
}

func NewReadable(gen types.Generator) (types.Readable, error) {
	r := &Stream{
		emitter:  emitter.NewEmitter(),
		state:    NewReadableState(),
		gen:      gen,
		doneChan: make(chan struct{}),
		Type:     types.ReadableType,
	}

	return r, nil
}
