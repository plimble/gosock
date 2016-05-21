package gosock

type hub struct {
	room map[string]*Room

	connections map[string]map[int64]*Connection

	defaultRoom map[*Connection]struct{}

	// Register requests from the connections.
	joinDefaultRoom chan *responseMessage
	join            chan *responseMessage

	// Unregister requests from connections.
	leave chan *responseMessage

	// Inbound messages from the connections.
	broadcast chan *responseMessage

	request chan *responseMessage

	to chan *responseMessage

	unregister chan *Connection
}

var h = &hub{
	room:            map[string]*Room{ROOM_DEFAULT: newRoom()},
	connections:     make(map[string]map[int64]*Connection),
	defaultRoom:     make(map[*Connection]struct{}),
	joinDefaultRoom: make(chan *responseMessage, 256),
	join:            make(chan *responseMessage, 256),
	leave:           make(chan *responseMessage, 256),
	broadcast:       make(chan *responseMessage, 256),
	request:         make(chan *responseMessage, 256),
	to:              make(chan *responseMessage, 256),
	unregister:      make(chan *Connection, 256),
}

func runHub() {
	for {
		select {
		case c := <-h.joinDefaultRoom:
			h.defaultRoom[c.connection] = struct{}{}
			if c.connection.uid != "" {
				if _, ok := h.connections[c.connection.uid]; ok {
					h.connections[c.connection.uid][c.connection.id] = c.connection
				} else {
					h.connections[c.connection.uid] = map[int64]*Connection{
						c.connection.id: c.connection,
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
		case c := <-h.leave:
			if _, ok := h.room[c.room]; ok {
				h.room[c.room].unregister(c.connection)
				if h.room[c.room].isEmptyRoom() {
					delete(h.room, c.room)
				}
			}
		case c := <-h.unregister:
			delete(h.defaultRoom, c)
			if _, ok := h.connections[c.uid]; ok {
				if _, ok := h.connections[c.uid][c.id]; ok {
					delete(h.connections[c.uid], c.id)
				}
			}

			for _, room := range h.room {
				if _, ok := room.connections[c]; ok {
					delete(room.connections, c)
				}
			}

			close(c.send)
			evt.emit(EVENT_DISCONNECTED, nil, nil)
		case c := <-h.broadcast:
			// pretty.Println(h.defaultRoom)
			if c.room == "" {
				for conn := range h.defaultRoom {
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
			} else {
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
			}
		case c := <-h.request:
			c.connection.send <- c.Data()
		case c := <-h.to:
			if _, ok := h.connections[c.to]; ok {
				for _, conn := range h.connections[c.to] {
					conn.send <- c.Data()
				}
			}
		}
	}
}
