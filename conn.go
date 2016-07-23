package gosock

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Connection struct {
	uid   string
	ws    *websocket.Conn
	send  chan []byte
	room  map[string]struct{}
	hub   *hub
	event *serverEventEmitter
}

func newConnection(uid string, c *websocket.Conn, hub *hub, event *serverEventEmitter) *Connection {
	return &Connection{
		uid:   uid,
		ws:    c,
		send:  make(chan []byte, 512),
		room:  make(map[string]struct{}),
		hub:   hub,
		event: event,
	}
}

func (c *Connection) readPump() {
	defer func() {
		pretty.Println("close")
		c.hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
				c.event.emitError(err, c)
			}
			break
		}

		pretty.Println(string(message))
		event, data := getRequestMessage(message)

		c.event.emit(event, data, c)
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	pretty.Println("writepump")
	for {
		select {
		case message, ok := <-c.send:
			pretty.Println(string(message))
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			pretty.Println("ping")
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Connection) register() {
	c.hub.register <- c
}

func (c *Connection) Join(room string) {
	c.hub.join <- newResponseMessage(room, "", "", nil, c)
}

func (c *Connection) Leave(room string) {
	c.hub.leave <- newResponseMessage(room, "", "", nil, c)
}

func (c *Connection) BroadcastRoom(room, event string, data []byte) {
	c.hub.broadcast <- newResponseMessage(room, event, "", data, c)
}

func (c *Connection) Broadcast(event string, data []byte) {
	c.hub.broadcast <- newResponseMessage(ROOM_DEFAULT, event, "", data, c)
}

func (c *Connection) Reply(event string, data []byte) {
	c.hub.request <- newResponseMessage(ROOM_DEFAULT, event, "", data, c)
}

func (c *Connection) To(id, event string, data []byte) {
	c.hub.to <- newResponseMessage(ROOM_DEFAULT, event, "", data, c)
}

func (c *Connection) Close() {
	c.Leave(ROOM_DEFAULT)
	c.hub.unregister <- c
	c.event.emit(EVENT_DISCONNECTED, nil, c)
}
