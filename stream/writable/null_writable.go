package writable

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

// NullStream type
type NullStream struct {
	emitter.Emitter
	Type     types.StreamType
	doneChan chan struct{}
}

// Pipe function
func (ws *NullStream) Pipe(w types.Writable) types.Writable {
	panic("cannot call pipe on writable stream")
}

func (ws *NullStream) Write(data types.Chunk) bool {
	if data == nil {
		go func() {
			close(ws.doneChan)
		}()
		return false
	}

	return true
}

func (ws *NullStream) Done() <-chan struct{} {
	return ws.doneChan
}

func NewNullWritable() (types.Writable, error) {
	ws := &NullStream{
		doneChan: make(chan struct{}),
	}

	return ws, nil
}
