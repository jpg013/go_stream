package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
)

type Chunk interface{}

const (
	// ReadableFlowing mode reads from the underlying system automatically
	// and provides to an application as quickly as possible.
	ReadableFlowing uint32 = 1
	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
	ReadableNotFlowing uint32 = 2
	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
	ReadableNull uint32 = 0
)

type ReadableConfig struct {
	highWaterMark uint
	generator     generators.Type
	read          func()
}

type StreamConfig struct {
	highWaterMark uint
	generator     generators.Type
	read          func()
	output        output.Type
}

func main() {
	// Define number generator
	gen, _ := generators.NewNumberGenerator(50)
	rs, err := NewReadable(StreamConfig{
		highWaterMark: 10,
		generator:     gen,
	})

	if err != nil {
		log.Fatal(err)
	}

	output, _ := output.NewSTDOUT(output.NewConfig())
	ws, err := NewWritableStream(StreamConfig{
		output:        output,
		highWaterMark: 50,
	})

	// mapOperator, _ := operators.NewMapOperator()

	if err != nil {
		log.Fatal(err)
	}

	rs.Pipe(ws)

	// Wait for read stream to finish
	<-ws.Done()
	time.Sleep(time.Second * 1)
	fmt.Println("Done")
}
