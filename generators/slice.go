package generators

import (
	"stream/types"
)

// SliceGenerator is an generator that takes in a slice and
// emits slice items individually as requested.
type SliceGenerator struct {
	slice []types.Chunk
	iter  int
}

func (sg *SliceGenerator) Next() (types.Chunk, error) {
	if sg.iter >= len(sg.slice) {
		return nil, nil
	}
	chunk := sg.slice[sg.iter]
	sg.iter++
	return chunk, nil
}

func NewSlice(slice []types.Chunk) (Type, error) {
	return &SliceGenerator{
		slice: slice,
		iter:  0,
	}, nil
}
