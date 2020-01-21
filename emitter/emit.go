package emitter

import (
	"sync/atomic"

	"github.com/jpg013/go_stream/types"
)

// Emit will push data to a specified topic
func (e *Emitter) Emit(topic string, data types.Chunk) {
	e.rw.RLock()

	fns, ok := e.handlers[topic]

	if !ok {
		e.rw.RUnlock()
		return
	}

	// this is done because the slices refer to same array even though they are passed by value
	// thus we are creating a new slice with our elements thus preserve locking correctly.
	handlers := append(EventHandlerSlice{}, fns...)
	e.rw.RUnlock()

	func(evt types.Event, handlers EventHandlerSlice) {
		for _, handler := range handlers {
			if handler.Once {
				if atomic.CompareAndSwapInt32(&handler.Fired, 0, 1) {
					// remove the handler from the topic
					e.removeHandler(handler, topic)
					handler.Callback(evt)
				}
			} else {
				handler.Callback(evt)
			}
		}
	}(types.Event{Topic: topic, Data: data}, handlers)
}
