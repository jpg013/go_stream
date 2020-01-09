package output

import "github.com/jpg013/go_stream/types"

type Type interface {
	Write(types.Chunk) error
	Close()
}
