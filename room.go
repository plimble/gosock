package gosock

const ROOM_DEFAULT = "_0"

type Room struct {
	connections map[*Connection]bool
}

func newRoom() *Room {
	return &Room{
		connections: make(map[*Connection]bool),
	}
}

func (r *Room) register(c *Connection) {
	r.connections[c] = true
}

func (r *Room) unregister(c *Connection) {
	delete(r.connections, c)
}

func (r *Room) isEmptyRoom() bool {
	if len(r.connections) > 0 {
		return false
	}

	return true
}
