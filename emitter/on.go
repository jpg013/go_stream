package emitter

import "stream/types"

// On subscribes an event channel to a particular topic
func (eb *Emitter) On(topic string, fn types.EventHandler) {
	eb.rw.Lock()

	if prev, ok := eb.handlers[topic]; ok {
		eb.handlers[topic] = append(prev, fn)
	} else {
		eb.handlers[topic] = []types.EventHandler{fn}
	}

	eb.rw.Unlock()
}
