package transform

import (
	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

// Stream type represents the base Transform stream
type Stream struct {
	emitter  *emitter.Type
	state    *TransformState
	Type     types.StreamType
	doneChan chan struct{}
}
