package main

import (
	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/stream/readable"
	"github.com/jpg013/go_stream/stream/writable"
)

func withHighWaterMark(hwm int) readable.OptionFunc {
	return func(c *readable.Config) {
		c.HighWaterMark = 16
	}
}

func withGenerator(size int) readable.OptionFunc {
	return func(c *readable.Config) {
		gen, _ := generators.NewNumberGenerator(size)
		c.Generator = gen
	}
}

func main() {
	rsConf := readable.NewConfig(
		withHighWaterMark(16),
		withGenerator(10),
	)

	wsConf := writable.NewConfig(func(c *writable.Config) {
		stdOut, _ := output.NewSTDOUT(output.NewConfig())
		c.Out = stdOut
	}, func(c *writable.Config) {
		c.HighWaterMark = 2
	})

	rs, _ := readable.NewReadableStream(rsConf)
	ws, _ := writable.NewWritableStream(wsConf)

	rs.Pipe(ws)

	// wait for write stream to finish
	<-ws.Done()
}
