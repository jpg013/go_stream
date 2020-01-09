package emitter

import (
	"stream/types"
	"sync"
)

// EventHandlerSlice is a slice of EventChannels
type EventHandlerSlice []types.EventHandler

// Emitter stores the information about listeners interested in a particular topic
type Emitter struct {
	handlers map[string]EventHandlerSlice
	rw       sync.RWMutex
}

// NewEmitter returns pointer to Emitter
func NewEmitter() *Emitter {
	return &Emitter{
		handlers: make(map[string]EventHandlerSlice),
	}
}
