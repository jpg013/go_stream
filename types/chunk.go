package types

type Chunk interface{}

type Mapper func(Chunk) (Chunk, error)
