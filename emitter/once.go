package emitter

import (
	"github.com/jpg013/go_stream/types"
)

// Once does a one time subscribe of an event handler to a particular topic
func (e *Emitter) Once(topic string, fn types.EventHandler) {
	e.rw.Lock()

	wrappedHandler := func(evt types.Event) {
		// Call Off, to remove the function event handler
		e.Off(topic, fn)

		// Call the original event handler
		fn(evt)
	}

	if prev, ok := e.handlers[topic]; ok {
		e.handlers[topic] = append(prev, wrappedHandler)
	} else {
		e.handlers[topic] = []types.EventHandler{wrappedHandler}
	}
	e.rw.Unlock()
}
