package writable

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
	"sync"
)

// Stream type
type Stream struct {
	mux sync.RWMutex
	emitter.Emitter
	state    *WritableState
	Type     types.StreamType
	doneChan chan struct{}
}
