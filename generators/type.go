package generators

import "stream/types"

type Type interface {
	Next() (types.Chunk, error)
}
