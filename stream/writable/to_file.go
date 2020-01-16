package writable

import (
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/types"
)

func ToFile(conf output.Config) (types.Writable, error) {
	output, err := output.NewFile(conf)

	if err != nil {
		return nil, err
	}

	return NewWritableStream(output)
}
