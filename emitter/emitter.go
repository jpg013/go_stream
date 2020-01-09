package emitter

import (
	"github.com/jpg013/go_stream/types"
	"sync"
)

// EventHandlerSlice is a slice of EventChannels
type EventHandlerSlice []types.EventHandler

// Type stores the information about listeners interested in a particular topic
type Type struct {
	handlers map[string]EventHandlerSlice
	rw       sync.RWMutex
}

// NewEmitter returns pointer to Emitter
func NewEmitter() *Type {
	return &Type{
		handlers: make(map[string]EventHandlerSlice),
	}
}
