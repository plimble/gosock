package gosocklite

type hub struct {
	connections map[*Connection]struct{}

	reply      chan *Connection
	publish    chan *Connection
	register   chan *Connection
	unregister chan *Connection
}

var h = &hub{
	connections: make(map[*Connection]struct{}),
	register:    make(chan *Connection, 256),
	unregister:  make(chan *Connection, 256),
}

func runHub() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = struct{}{}
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
			}

			close(c.send)
			evt.emit(EVENT_DISCONNECTED, nil, nil)
		}
	}
}
