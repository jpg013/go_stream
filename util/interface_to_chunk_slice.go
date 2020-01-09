package util

import (
	"errors"
	"github.com/jpg013/go_stream/types"
	"reflect"
)

// InterfaceToChunkSlice tries to convert an empty interface{} to a
// slice of Chunks, or returns an error if unsuccessful.
func InterfaceToChunkSlice(slice interface{}) ([]types.Chunk, error) {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		return nil, errors.New("InterfaceToChunkSlice() given a non-slice type")
	}

	ret := make([]types.Chunk, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, nil
}
