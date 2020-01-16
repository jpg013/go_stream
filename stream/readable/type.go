package readable

import (
	"sync"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

type ReadableMode int32

const (
	// ReadableFlowing mode reads from the underlying system automatically
	// and provides to an application as quickly as possible.
	ReadableFlowing int32 = 1
	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
	ReadableNotFlowing int32 = 2
	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
	ReadableNull int32 = 0
)

// Stream represents a readable stream type
type Stream struct {
	// embedded event emitter
	emitter.Emitter
	state *ReadableState
	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest       types.Writable
	mux        sync.RWMutex
	doneChan   chan struct{}
	StreamType types.StreamType
	read       func()
}
