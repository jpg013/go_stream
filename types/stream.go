package types

type StreamType uint32

const (
	ReadableType StreamType = iota
	DuplexType
	WritableType
)

type Stream interface {
	Pipe(Stream) Stream
	On(string, EventHandler)
	Emit(string, interface{})
}
