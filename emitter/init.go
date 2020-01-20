package emitter

// Emit will push data to a specified topic
func (e *Emitter) Init() {
	e.handlers = make(map[string]EventHandlerSlice)
}
