package writable

import (
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/types"
)

func ToSTDOUT(conf output.Config) (types.Writable, error) {
	output, err := output.NewSTDOUT(conf)

	if err != nil {
		return nil, err
	}

	return NewWritable(output), nil
}
