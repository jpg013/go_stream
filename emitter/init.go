package emitter

// Emit will push data to a specified topic
func (e *Emitter) Init() {
	e.rw.RLock()
	defer e.rw.RUnlock()

	e.handlers = make(map[string]EventHandlerSlice)
}
