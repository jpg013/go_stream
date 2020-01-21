package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/operators"
	"github.com/jpg013/go_stream/output"
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

func withSTDOUT() stream.OptionFunc {
	return func(c *stream.Config) {
		out, _ := output.NewSTDOUT(output.NewConfig())
		c.Writable.Out = out
	}
}

func main() {
	conf := stream.NewConfig(
		withTransformOperator(),
		withReadableSource(),
		withSTDOUT(),
	)

	r, _ := stream.NewReadableStream(conf)
	t, _ := stream.NewTransformStream(conf)
	w, _ := stream.NewWritableStream(conf)

	r.Pipe(t).Pipe(w)

	// Wait for stream to end
	<-w.Done()
}
