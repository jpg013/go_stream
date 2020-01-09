package readable

import (
	"stream/generators"
	"stream/util"
)

func FromSlice(data interface{}) (Readable, error) {
	slice, err := util.InterfaceToChunkSlice(data)

	if err != nil {
		return nil, err
	}

	gen, err := generators.NewSlice(slice)

	if err != nil {
		return nil, err
	}

	return NewReadable(gen), nil
}
