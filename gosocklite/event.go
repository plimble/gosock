package gosocklite

const (
	EVENT_CONNECTION   = "connection"
	EVENT_DISCONNECTED = "disconnected"
	EVENT_CLOSED       = "closed"
)

type EventHandler func([]byte, *Connection)

type eventEmitter struct {
	events map[string][]EventHandler
}

var evt = &eventEmitter{
	events: make(map[string][]EventHandler),
}

func (e *eventEmitter) emit(event string, data []byte, c *Connection) {
	if _, ok := e.events[event]; ok {
		for _, handler := range e.events[event] {
			handler(data, c)
		}
	}
}

func (e *eventEmitter) register(event string, handler EventHandler) {
	if _, ok := e.events[event]; ok {
		e.events[event] = append(e.events[event], handler)
	} else {
		e.events[event] = []EventHandler{handler}
	}
}
