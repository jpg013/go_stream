package generators

import "github.com/jpg013/go_stream/types"

type Type interface {
	Next() (types.Chunk, error)
}
