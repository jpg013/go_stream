package emitter

import "github.com/jpg013/go_stream/types"

// On subscribes an event channel to a particular topic
func (e *Emitter) On(topic string, fn func(types.Event)) func() {
	e.rw.Lock()

	// If handlers haven't been setup, the Init()
	if e.handlers == nil {
		e.Init()
	}

	handler := e.newHandler(fn, false)

	if prev, ok := e.handlers[topic]; ok {
		e.handlers[topic] = append(prev, handler)
	} else {
		e.handlers[topic] = []*types.EventHandler{handler}
	}

	e.rw.Unlock()

	return func() {
		e.removeHandler(handler, topic)
	}
}
