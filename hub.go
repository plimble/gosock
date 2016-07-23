package gosock

type hub struct {
	room map[string]*room

	connections map[string]map[*Connection]struct{}

	// Register requests from the connections.
	joinDefaultRoom chan *responseMessage
	join            chan *responseMessage

	// Unregister requests from connections.
	leave chan *responseMessage

	// Inbound messages from the connections.
	broadcast chan *responseMessage

	request chan *responseMessage

	to chan *responseMessage

	register   chan *Connection
	unregister chan *Connection
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			if c.uid != "" {
				if _, ok := h.connections[c.uid]; ok {
					h.connections[c.uid][c] = struct{}{}
				} else {
					h.connections[c.uid] = map[*Connection]struct{}{
						c: struct{}{},
					}
				}
			}
		case c := <-h.join:
			if _, ok := h.room[c.room]; ok {
				h.room[c.room].register(c.connection)
			} else {
				h.room[c.room] = newRoom()
				h.room[c.room].register(c.connection)
			}

			for conn := range h.connections[c.connection.uid] {
				if _, ok := conn.room[c.room]; !ok {
					conn.room[c.room] = struct{}{}
				}
			}
		case c := <-h.leave:
			if _, ok := h.room[c.room]; ok {
				h.room[c.room].unregister(c.connection)
				if h.room[c.room].isEmptyRoom() {
					delete(h.room, c.room)
				}
			}

			for conn := range h.connections[c.connection.uid] {
				if _, ok := conn.room[c.room]; ok {
					delete(conn.room, c.room)
				}
			}
		case c := <-h.unregister:
			// delete(h.defaultRoom, c)
			if _, ok := h.connections[c.uid]; ok {
				if _, ok := h.connections[c.uid][c]; ok {
					delete(h.connections[c.uid], c)
				}
			}

			close(c.send)
		case c := <-h.broadcast:
			for conn := range h.room[c.room].connections {
				if conn == c.connection {
					continue
				}
				select {
				case conn.send <- c.Data():
				default:
					h.unregister <- c.connection
					close(conn.send)
				}
			}
		case c := <-h.request:
			c.connection.send <- c.Data()
		case c := <-h.to:
			if _, ok := h.connections[c.to]; ok {
				for conn := range h.connections[c.to] {
					conn.send <- c.Data()
				}
			}
		}
	}
}
