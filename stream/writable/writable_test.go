package writable

import (
	"testing"

	"github.com/jpg013/go_stream/generators"
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/stream/readable"
)

func BenchmarkWritable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out, _ := output.NewNullOutput()
		ws, _ := NewWritable(out)
		gen, _ := generators.NewNumberGenerator(1000)
		rs, _ := readable.NewReadable(gen)

		rs.Pipe(ws)

		// wait for writable stream to finish
		<-ws.Done()
	}
}