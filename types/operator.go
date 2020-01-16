package types

type Operator interface {
	Exec(Chunk)
	End()
	GetOutput() <-chan Chunk
	GetError() <-chan error
}
