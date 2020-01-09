package output

import "stream/types"

type Type interface {
	Write(types.Chunk) error
	Close()
}
