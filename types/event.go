package types

type EventHandler struct {
	Callback func(Event)
	ID       int
	Fired    int32
	Once     bool
}

// Event represents a data structure containing data and topic that can
// be passed to an event channel
type Event struct {
	Data  interface{}
	Topic string
}
