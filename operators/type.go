package operators

import "github.com/jpg013/go_stream/types"

type Type interface {
	Exec(types.Chunk) error
	End()
	GetOutput() <-chan types.Chunk
	GetError() <-chan error
}

type Mapper func(types.Chunk) (types.Chunk, error)
