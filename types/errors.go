package types

import (
	"errors"
)

var (
	ErrAlreadyStarted = errors.New("type has already been started")
	ErrOutputClosed   = errors.New("output has been closed and cannot be written to")
)
