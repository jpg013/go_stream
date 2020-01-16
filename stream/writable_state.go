package stream

// import (
// 	"container/list"
// 	"sync"
// )

// type WritableState struct {
// 	buffer        *list.List
// 	highWaterMark int
// 	draining      bool
// 	writing       bool
// 	ended         bool
// 	destroyed     bool
// 	mux           sync.Mutex
// }

// func NewWritableState() *WritableState {
// 	return &WritableState{
// 		buffer:        list.New(),
// 		highWaterMark: 3,
// 		writing:       false,
// 		draining:      false,
// 		ended:         false,
// 		destroyed:     false,
// 	}
// }
