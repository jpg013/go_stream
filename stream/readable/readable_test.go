package readable

import (
	"testing"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/stream/writable"
)

func BenchmarkReadable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ws, _ := writable.NewNullWritable()
		gen, _ := generators.NewNumberGenerator(2)
		rs, _ := NewReadable(gen)

		rs.Pipe(ws)

		// wait for writeble stream to finish
		<-ws.Done()
	}
}
