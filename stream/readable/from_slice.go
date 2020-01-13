package readable

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/types"
	"github.com/jpg013/go_stream/util"
)

func FromSlice(data interface{}) (types.Readable, error) {
	slice, err := util.InterfaceToChunkSlice(data)

	if err != nil {
		return nil, err
	}

	gen, err := generators.NewSliceGenerator(slice)

	if err != nil {
		return nil, err
	}

	return NewReadable(gen)
}
