package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/operators"
	"github.com/jpg013/go_stream/stream"
	"github.com/jpg013/go_stream/stream/readable"
	"github.com/jpg013/go_stream/types"
)

func main() {
	mapper := func(chunk types.Chunk) (types.Chunk, error) {
		return chunk.(int) * 2, nil
	}
	operator, _ := operators.NewMapOperator(mapper)
	gen, _ := generators.NewNumberGenerator(10)
	rs, _ := readable.NewReadableStream(gen)
	ts, _ := stream.NewTransformStream(operator)

	rs.Pipe(ts) //.Pipe(ws)

	// Wait for stream to end
	<-ts.Done()
}
