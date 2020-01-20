package types

type StreamType uint32

const (
	ReadableType StreamType = iota
	TransformType
	WritableType
)

type Stream interface {
	EventEmitter
	Pipe(Writable) Writable
	Done() <-chan struct{}
}

type Reader interface {
	Read() Chunk
}

type Writer interface {
	Write(Chunk) bool
}

type Readable interface {
	Stream
	Reader
}

type Writable interface {
	Stream
	Writer
}

type Transform interface {
	Stream
	Reader
	Writer
}
