package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/operators"
	"github.com/jpg013/go_stream/stream"
	"github.com/jpg013/go_stream/types"
)

func double(c types.Chunk) (types.Chunk, error) {
	return c.(int) * 2, nil
}

func withTransformOperator() stream.OptionFunc {
	return func(c *stream.Config) {
		operator, _ := operators.NewMapOperator(double)
		c.Transform.Operator = operator
	}
}

func withReadableSource() stream.OptionFunc {
	return func(c *stream.Config) {
		gen, _ := generators.NewNumberGenerator(10)
		c.Readable.Generator = gen
	}
}

func main() {
	conf := stream.NewConfig(
		withTransformOperator(),
		withReadableSource(),
	)

	rs, _ := stream.NewReadableStream(conf)
	ts, _ := stream.NewTransformStream(conf)

	rs.Pipe(ts) //.Pipe(ws)

	// Wait for stream to end
	<-ts.Done()
}
