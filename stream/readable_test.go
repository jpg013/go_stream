package stream

import (
	"testing"

	"github.com/jpg013/go_stream/generators"
)

func BenchmarkReadable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ws, _ := NewNullWritable()
		rs, _ := newNumberReadable()

		rs.Pipe(ws)

		// wait for writeble stream to finish
		<-ws.Done()
	}
}

func newNumberReadable() (Readable, error) {
	gen, _ := generators.NewNumberGenerator(100)
	config := NewConfig(func(c *Config) {
		c.Readable.Generator = gen
	})
	return NewReadableStream(config)
}
