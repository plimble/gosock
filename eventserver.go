package gosock

const (
	EVENT_CONNECTION   = "connection"
	EVENT_DISCONNECTED = "disconnected"
)

type ServerEventHandler func([]byte, *Connection)
type ServerErrorHandler func(error, *Connection)

type serverEventEmitter struct {
	events map[string][]ServerEventHandler
	errors []ServerErrorHandler
}

func (e *serverEventEmitter) emit(event string, data []byte, c *Connection) {
	if _, ok := e.events[event]; ok {
		for _, handler := range e.events[event] {
			handler(data, c)
		}
	}
}

func (e *serverEventEmitter) emitError(err error, c *Connection) {
	for _, handler := range e.errors {
		handler(err, c)
	}
}

func (e *serverEventEmitter) register(event string, handler ServerEventHandler) {
	if _, ok := e.events[event]; ok {
		e.events[event] = append(e.events[event], handler)
	} else {
		e.events[event] = []ServerEventHandler{handler}
	}
}

func (e *serverEventEmitter) registerError(handler ServerErrorHandler) {
	e.errors = append(e.errors, handler)
}
