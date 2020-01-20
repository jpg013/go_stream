package emitter

import (
	"reflect"
	"sync"

	"github.com/jpg013/go_stream/types"
)

// EventHandlerSlice is a slice of EventChannels
type EventHandlerSlice []*types.EventHandler

// Emitter stores the information about listeners interested in a particular topic
type Emitter struct {
	handlers  map[string]EventHandlerSlice
	rw        sync.RWMutex
	idCounter int
}

// NewEmitter returns pointer to Emitter
func NewEmitter() *Emitter {
	return &Emitter{
		handlers:  make(map[string]EventHandlerSlice),
		idCounter: 0,
	}
}

func compareCallbacks(a func(types.Event), b func(types.Event)) bool {
	aAddr := reflect.ValueOf(a).Pointer()
	bAddr := reflect.ValueOf(b).Pointer()

	return aAddr == bAddr
}

// compareHandlers compares function pointers and returns true if pointer
// values are the same
func compareHandlers(a *types.EventHandler, b *types.EventHandler) bool {
	return a.ID == b.ID
}

func (e *Emitter) removeHandler(item *types.EventHandler, topic string) {
	e.rw.Lock()
	if handlers, ok := e.handlers[topic]; ok {
		tmp := handlers[:0]
		for _, handler := range handlers {
			if !compareHandlers(item, handler) {
				tmp = append(tmp, handler)
			}
		}
		e.handlers[topic] = handlers
	}
	e.rw.Unlock()
}

func (e *Emitter) newHandler(cb func(types.Event), once bool) *types.EventHandler {
	e.idCounter++

	handler := &types.EventHandler{
		Fired:    0,
		Callback: cb,
		ID:       e.idCounter,
		Once:     once,
	}

	return handler
}
