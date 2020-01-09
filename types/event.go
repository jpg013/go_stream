package types

type EventHandler func(Event)

// Event represents a data structure containing data and topic that can
// be passed to an event channel
type Event struct {
	Data  interface{}
	Topic string
}
