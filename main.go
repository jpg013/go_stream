package main

import (
	"fmt"
	"log"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
)

type Chunk interface{}

type ReadableConfig struct {
	highWaterMark uint
	generator     generators.Type
	read          func()
}

func main() {
	// Define number generator
	gen, _ := generators.NewNumberGenerator(7500)

	rs, err := NewReadable(ReadableConfig{
		highWaterMark: 99,
		generator:     gen,
	})

	if err != nil {
		log.Fatal(err)
	}

	output, _ := output.NewSTDOUT(output.NewConfig())

	ws, err := NewWritableStream(WritableConfig{
		output:        output,
		highWaterMark: 50,
	})

	if err != nil {
		log.Fatal(err)
	}

	rs.Pipe(ws)

	// Wait for read stream to finish
	<-ws.Done()
	fmt.Println("Done")
}
