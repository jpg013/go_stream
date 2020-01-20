package stream

import (
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func fromReadableBuffer(state *ReadableState) types.Chunk {
	state.mux.RLock()
	defer state.mux.RUnlock()

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
