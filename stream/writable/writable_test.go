package writable

// func BenchmarkWritable(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		out, _ := output.NewNullOutput()
// 		ws, _ := NewWritableStream(out)
// 		gen, _ := generators.NewNumberGenerator(1000)
// 		rs, _ := readable.NewReadableStream(gen)

// 		rs.Pipe(ws)

// 		// wait for writable stream to finish
// 		<-ws.Done()
// 	}
// }
