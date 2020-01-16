package main

import (
	"fmt"
	"time"

	"github.com/jpg013/go_stream/emitter"
	"github.com/jpg013/go_stream/types"
)

func main() {
	emitter := emitter.NewEmitter()

	handler := func(evt types.Event) {
		fmt.Println(evt.Data)
	}

	emitter.On("test", handler)
	emitter.On("test", func(evt types.Event) {
		fmt.Println("in the other handler!!")
	})
	emitter.Emit("test", "justin")
	time.Sleep(time.Second * 1)
	emitter.Off("test", handler)
	emitter.Emit("test", "justin")
	time.Sleep(time.Second * 1)
}
