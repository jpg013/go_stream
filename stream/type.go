package stream

// type ReadableMode uint32

// type StreamType uint32

// const (
// 	// ReadableFlowing mode reads from the underlying system automatically
// 	// and provides to an application as quickly as possible.
// 	ReadableFlowing ReadableMode = iota
// 	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
// 	ReadableNotFlowing
// 	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
// 	ReadableNull
// )

// const (
// 	ReadableType StreamType = iota
// 	DuplexType
// 	WritableType
// )

// type Chunk interface{}

// // type Operator func(data interface{})

// // Sink is an operator with exactly one input, requesting and accepting data elements,
// // possible slowing down the upstream producer elements.
// // type Sink interface {
// // }

// // DataSource is an operator with exactly one output, emitting data elements
// // whenever downstream operators are ready to receive them.
// // type DataSource interface {
// // 	ReadStart() (interface{}, error)
// // 	ReadStop() error
// // }

// type DataSource interface {
// 	Next() (Chunk, error)
// }

// type Writer interface {
// 	DoWrite(Chunk) error
// 	End()
// }

// type Stream interface {
// 	Pipe(Writable) Stream
// 	On(string, EventHandler)
// 	Emit(string, interface{})
// }

// type Readable interface {
// 	Stream
// 	Read() bool
// }

// type Writable interface {
// 	Stream
// 	Write(Chunk) bool
// }

// // type Executor func(interface{}) (interface{}, error)
