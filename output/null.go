package output

import (
	"github.com/jpg013/go_stream/types"
)

type Null struct {
}

func (n *Null) Write(chunk types.Chunk) error {
	// PASS
	return nil
}

func (n *Null) Close() {
	// Pass
}

// NewNullOutput returns a Null output type.
func NewNullOutput() (Type, error) {
	return &Null{}, nil
}
