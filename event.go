package gosock

const (
	EVENT_CONNECTION   = "connection"
	EVENT_DISCONNECTED = "disconnected"
	EVENT_CLOSED       = "closed"
)

type EventHandler func([]byte, *Connection)
type ErrorHandler func(error, *Connection)

type eventEmitter struct {
	events map[string][]EventHandler
	errors []ErrorHandler
}

var evt = &eventEmitter{
	events: make(map[string][]EventHandler),
	errors: []ErrorHandler{},
}

func (e *eventEmitter) emit(event string, data []byte, c *Connection) {
	if _, ok := e.events[event]; ok {
		for _, handler := range e.events[event] {
			handler(data, c)
		}
	}
}

func (e *eventEmitter) emitError(err error, c *Connection) {
	for _, handler := range e.errors {
		handler(err, c)
	}
}

func (e *eventEmitter) register(event string, handler EventHandler) {
	if _, ok := e.events[event]; ok {
		e.events[event] = append(e.events[event], handler)
	} else {
		e.events[event] = []EventHandler{handler}
	}
}

func (e *eventEmitter) registerError(handler ErrorHandler) {
	e.errors = append(e.errors, handler)
}
