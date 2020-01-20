package types

type EventEmitter interface {
	Emit(string, Chunk)
	On(string, func(Event)) func()
	Once(string, func(Event)) func()
}
