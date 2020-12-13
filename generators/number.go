package generators

import (
	"github.com/jpg013/go_stream/types"
)

// NumberGenerator generates an incremented integer each time.
type NumberGenerator struct {
	iter int
	max  int
}

func (g *NumberGenerator) Next() (types.Chunk, error) {
	if g.iter > g.max {
		return nil, nil
	}
	chunk := g.iter
	g.iter++
	return chunk, nil
}

func NewNumberGenerator(max int) (Type, error) {
	return &NumberGenerator{
		iter: 1,
		max:  max,
	}, nil
}
