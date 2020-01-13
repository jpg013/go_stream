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

// NewNull creates a Null output type.
func NewNull() (Type, error) {
	return &Null{}, nil
}
