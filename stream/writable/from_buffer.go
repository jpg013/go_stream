package writable

import (
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

func fromBuffer(ws *Stream) types.Chunk {
	state := ws.state
	ws.mux.Lock()
	defer ws.mux.Unlock()

	if getLength(ws) == 0 {
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
