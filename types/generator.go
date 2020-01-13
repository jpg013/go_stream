package types

type Generator interface {
	Next() (Chunk, error)
}
