package types

type StreamType uint32

const (
	ReadableType StreamType = iota
	TransformType
	WritableType
)

type EventEmitter interface {
	On(string, EventHandler)
	Emit(string, interface{})
}

type Stream interface {
	EventEmitter
	Pipe(Writable) Writable
	Done() <-chan struct{}
}

type Readable interface {
	Stream
	Read() bool
}

type Writable interface {
	Stream
	Write(Chunk) bool
}

type Transform interface {
	Stream
	Read() bool
	Write(Chunk) bool
	Transform(Chunk) bool
}
