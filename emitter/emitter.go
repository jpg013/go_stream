package emitter

import (
	"reflect"
	"sync"

	"github.com/jpg013/go_stream/types"
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

// compareHandlers compares function pointers and returns true if pointer
// values are the same
func compareHandlers(a types.EventHandler, b types.EventHandler) bool {
	aAddr := reflect.ValueOf(a).Pointer()
	bAddr := reflect.ValueOf(b).Pointer()

	return aAddr == bAddr
}

func removeHandler(fn types.EventHandler, slice EventHandlerSlice) EventHandlerSlice {
	tmp := slice[:0]
	for _, handler := range slice {
		if !compareHandlers(fn, handler) {
			tmp = append(tmp, handler)
		}
	}
	return tmp
}
