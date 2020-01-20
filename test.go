package main

import (
	"sync"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

func makeCallback() func(types.Event) {
	mux := sync.Mutex{}
	count := 0

	return func(types.Event) {
		mux.Lock()
		count++
		mux.Unlock()
	}
}

func main() {
	emitter := emitter.NewEmitter()

	for i := 0; i < 10; i++ {
		emitter.Once("balls", makeCallback())
		emitter.Emit("balls", nil)
		emitter.Emit("balls", nil)
	}
}
