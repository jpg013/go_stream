package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/stream"
)

func withReadableHighWaterMark(hwm int) stream.OptionFunc {
	return func(c *stream.Config) {
		c.Readable.HighWaterMark = hwm
	}
}

func withWritableHighWaterMark(hwm int) stream.OptionFunc {
	return func(c *stream.Config) {
		c.Writable.HighWaterMark = hwm
	}
}

func withGenerator(size int) stream.OptionFunc {
	return func(c *stream.Config) {
		gen, _ := generators.NewNumberGenerator(size)
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
	config := stream.NewConfig(
		withReadableHighWaterMark(16),
		withWritableHighWaterMark(2),
		withGenerator(10),
		withSTDOUT(),
	)

	rs, _ := stream.NewReadableStream(config)
	ws, _ := stream.NewWritableStream(config)

	rs.Pipe(ws)
	// wait for write stream to finish
	<-ws.Done()
}
