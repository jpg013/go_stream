package emitter

import "github.com/jpg013/go_stream/types"

// Emit will push data to a specified topic
func (e *Emitter) Emit(topic string, data interface{}) {
	e.rw.RLock()
	defer e.rw.RUnlock()

	if fns, ok := e.handlers[topic]; ok {
		// this is done because the slices refer to same array even though they are passed by value
		// thus we are creating a new slice with our elements thus preserve locking correctly.
		handlers := append(EventHandlerSlice{}, fns...)
		func(evt types.Event, handlers EventHandlerSlice) {
			for _, fn := range handlers {
				go fn(evt)
			}
		}(types.Event{Topic: topic, Data: data}, handlers)
	}
}
