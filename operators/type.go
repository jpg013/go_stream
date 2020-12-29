package operators

import "github.com/jpg013/go_stream/types"

type Type interface {
	Exec(types.Chunk) error
	End()
	Output() <-chan types.Chunk
	Error() <-chan error
}

type Mapper func(types.Chunk) (types.Chunk, error)
