package readable

import (
	"testing"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/stream/writable"
	"github.com/jpg013/go_stream/types"
)

func BenchmarkReadable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ws, _ := writable.NewNullWritable()
		rs, _ := newNumberReadable()

		rs.Pipe(ws)

		// wait for writeble stream to finish
		<-ws.Done()
	}
}

func newNumberReadable() (types.Readable, error) {
	gen, _ := generators.NewNumberGenerator(2)
	config := NewConfig(func(c *Config) {
		c.generator = gen
	})
	return NewReadableStream(config)
}
