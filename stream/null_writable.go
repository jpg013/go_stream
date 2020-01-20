package stream

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

// NullStream type
type NullStream struct {
	emitter.Emitter
	Type     StreamType
	state    *WritableState
	doneChan chan struct{}
}

// Pipe function
func (ws *NullStream) Pipe(w Writable) Writable {
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

func (ws *NullStream) GetWriteState() *WritableState {
	return ws.state
}

func NewNullWritable() (Writable, error) {
	w := &NullStream{
		doneChan: make(chan struct{}),
		state:    &WritableState{},
	}

	return w, nil
}
