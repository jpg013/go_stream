package readable

import "container/list"

type ReadableState struct {
	buffer            *list.List
	mode              uint32
	pendingReads      int32
	length            int32
	reading           uint32
	highWaterMark     int
	ended             bool
	destroyed         bool
	awaitDrainWriters uint32
}

func NewReadableState() *ReadableState {
	return &ReadableState{
		buffer:            list.New(),
		mode:              ReadableNull,
		length:            0,
		pendingReads:      0,
		highWaterMark:     5,
		destroyed:         false,
		reading:           0,
		ended:             false,
		awaitDrainWriters: 0,
	}
}
