package main

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/types"
)

type Chunk interface{}

type Writable interface {
}

type Readable interface {
	Pipe(Writable) Writable
	Done() <-chan struct{}
	Read() types.Chunk
}

type ReadableStream struct {
	mux           sync.RWMutex
	buffer        *list.List
	mode          uint32
	highWaterMark int
	destroyed     bool

	// Destination for readable to write data,
	// this is set when Pipe() is called
	dest Writable

	// internal read method that can be overwritten
	read func()
}

type ReadableConfig struct {
	highWaterMark uint
	generator     generators.Type
}

func NewReadable(conf ReadableConfig) Readable {
	return nil
}

func main() {
	gen, _ := generators.NewNumberGenerator(100)

	rs := NewReadable(ReadableConfig{
		highWaterMark: 50,
		generator:     gen,
	})

	fmt.Println(rs)
}
