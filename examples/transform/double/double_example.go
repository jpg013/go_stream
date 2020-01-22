package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/stream"
	"github.com/jpg013/go_stream/types"
)

func double(c types.Chunk) (types.Chunk, error) {
	return c.(int) * 2, nil
}

func withTransformOperator() stream.OptionFunc {
	return func(c *stream.Config) {
		c.Transform.Mapper = func(chunk types.Chunk) (types.Chunk, error) {
			return chunk.(int) * 2, nil
		}
	}
}

func withReadableSource() stream.OptionFunc {
	return func(c *stream.Config) {
		gen, _ := generators.NewNumberGenerator(3)
		c.Readable.Generator = gen
	}
}

func withSTDOUT() stream.OptionFunc {
	return func(c *stream.Config) {
		out, _ := output.NewSTDOUT(output.NewConfig())
		c.Writable.Out = out
	}
}

func getReadable(conf *stream.Config) stream.Readable {
	r, _ := stream.NewReadableStream(conf)

	return r
}

func getTransform(conf *stream.Config) stream.Transform {
	t, _ := stream.NewTransformStream(conf)

	return t
}

func getWritable(conf *stream.Config) stream.Writable {
	w, _ := stream.NewWritableStream(conf)

	return w
}

func main() {
	for i := 0; i < 1; i++ {
		conf := stream.NewConfig(
			withTransformOperator(),
			withReadableSource(),
			// withSTDOUT(),
		)

		r := getReadable(conf)
		t := getTransform(conf)

		r.Pipe(t)

		<-t.Done()
	}
}
