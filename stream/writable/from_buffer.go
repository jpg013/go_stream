package writable

import (
	"container/list"
	"github.com/jpg013/go_stream/types"
)

func fromBuffer(buf *list.List) types.Chunk {
	val := buf.Front()

	if val == nil {
		return val
	}

	buf.Remove(val)

	return val.Value
}
