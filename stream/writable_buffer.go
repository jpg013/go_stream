package stream

import (
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func writableBufferChunk(w Writable, data types.Chunk) {
	state := w.GetWriteState()
	state.mux.Lock()
	defer state.mux.Unlock()

	state.buffer.PushBack(data)
	len := atomic.AddInt32(&state.length, 1)

	if int(len) >= state.highWaterMark && atomic.LoadUint32(&state.draining) == 0 {
		atomic.CompareAndSwapUint32(&state.draining, 0, 1)
	}
}

func unshiftWritableBuffer(state *WritableState) {

}

func shiftWritableBuffer(w Writable) types.Chunk {
	state := w.GetWriteState()
	state.mux.Lock()
	defer state.mux.Unlock()

	len := atomic.LoadInt32(&state.length)

	if len == 0 {
		return nil
	}

	val := state.buffer.Front()

	if val == nil {
		return val
	}

	state.buffer.Remove(val)

	atomic.AddInt32(&state.length, -1)

	return val.Value
}
