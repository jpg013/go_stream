package emitter

import "github.com/jpg013/go_stream/types"

// On subscribes an event channel to a particular topic
func (e *Type) On(topic string, fn types.EventHandler) {
	e.rw.Lock()

	if prev, ok := e.handlers[topic]; ok {
		e.handlers[topic] = append(prev, fn)
	} else {
		e.handlers[topic] = []types.EventHandler{fn}
	}

	e.rw.Unlock()
}
