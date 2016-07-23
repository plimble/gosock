package gosock

import (
	"sync"
)

const ROOM_DEFAULT = "_default"

type room struct {
	sync.RWMutex
	connections map[*Connection]struct{}
}

func newRoom() *room {
	return &room{
		connections: make(map[*Connection]struct{}),
	}
}

func (r *room) register(c *Connection) {
	r.Lock()
	r.connections[c] = struct{}{}
	r.Unlock()
}

func (r *room) unregister(c *Connection) {
	r.Lock()
	delete(r.connections, c)
	r.Unlock()
}

func (r *room) isEmptyRoom() bool {
	r.RLock()
	defer r.RUnlock()

	if len(r.connections) > 0 {
		return false
	}

	return true
}
