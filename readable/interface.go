package readable

import (
	"stream/types"
	"stream/writable"
)

type Readable interface {
	Pipe(writable.Writable) writable.Writable
	On(string, types.EventHandler)
	Emit(string, interface{})
	Read() bool
}
