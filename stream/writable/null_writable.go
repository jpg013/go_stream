package writable

import (
	"github.com/jpg013/go_stream/types"
)

// NullStream type
type NullStream struct {
	Type     types.StreamType
	doneChan chan struct{}
}

// On function
func (ws *NullStream) On(topic string, fn types.EventHandler) {
}

// Emit function
func (ws *NullStream) Emit(topic string, data interface{}) {

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
