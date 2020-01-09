package writable

import (
	"stream/types"
)

type Writer interface {
	DoWrite(types.Chunk) error
	End()
}

type Writable interface {
	Pipe(Writable) Writable
	Write(types.Chunk) bool
	On(string, types.EventHandler)
	Emit(string, interface{})
}
