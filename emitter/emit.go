package emitter

import "stream/types"

// Emit will push data to a specified topic
func (eb *Emitter) Emit(topic string, data interface{}) {
	eb.rw.RLock()

	if fns, ok := eb.handlers[topic]; ok {
		// this is done because the slices refer to same array even though they are passed by value
		// thus we are creating a new slice with our elements thus preserve locking correctly.
		handlers := append(EventHandlerSlice{}, fns...)
		go func(evt types.Event, handlers EventHandlerSlice) {
			for _, fn := range handlers {
				fn(evt)
			}
		}(types.Event{Topic: topic, Data: data}, handlers)
	}
	eb.rw.RUnlock()
}
