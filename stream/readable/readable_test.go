package readable

import "testing"

func BenchmarkReadable(b *testing.B) {
	// numSlice := []int{1, 2, 3, 4, 5}
	// rs, _ := FromSlice(numSlice)
	// ws, _ := writable.NewNullWritable()

	// for i := 0; i < b.N; i++ {
	// 	ws, _ := writable.NewNullWritable()
	// 	gen, _ := generators.NewNumberGenerator(1000)
	// 	rs, _ := NewReadableStream(gen)

	// 	rs.Pipe(ws)

	// 	// wait for writeble stream to finish
	// 	<-ws.Done()
	// }
}
