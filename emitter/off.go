package emitter

import "github.com/jpg013/go_stream/types"

// Off unsubscribes an event handler from a topic
func (e *Emitter) Off(topic string, fn types.EventHandler) {
	e.rw.Lock()

	handlers, ok := e.handlers[topic]

	if !ok {
		return
	}

	e.handlers[topic] = removeHandler(fn, handlers)
	e.rw.Unlock()
}
