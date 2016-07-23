package gosock

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type GoSock struct {
	hub   *hub
	event *serverEventEmitter
}

func New() *GoSock {
	return &GoSock{
		hub: &hub{
			room:        map[string]*room{ROOM_DEFAULT: newRoom()},
			connections: make(map[string]map[*Connection]struct{}),
			join:        make(chan *responseMessage, 1024),
			leave:       make(chan *responseMessage, 1024),
			broadcast:   make(chan *responseMessage, 1024),
			request:     make(chan *responseMessage, 1024),
			to:          make(chan *responseMessage, 1024),
			register:    make(chan *Connection, 1024),
			unregister:  make(chan *Connection, 1024),
		},
		event: &serverEventEmitter{
			events: make(map[string][]ServerEventHandler),
			errors: []ServerErrorHandler{},
		},
	}
}

func (g *GoSock) Handler() http.HandlerFunc {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	go g.hub.run()
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		c := newConnection(r.URL.Query().Get("uid"), ws, g.hub, g.event)
		c.register()
		c.Join(ROOM_DEFAULT)
		g.event.emit(EVENT_CONNECTION, nil, c)
		go c.writePump()
		c.readPump()
	}
}

func (g *GoSock) On(event string, handler ServerEventHandler) {
	g.event.register(event, handler)
}

func (g *GoSock) OnConnection(handler ServerEventHandler) {
	g.event.register(EVENT_CONNECTION, handler)
}

func (g *GoSock) OnError(handler ServerErrorHandler) {
	g.event.registerError(handler)
}

func (g *GoSock) OnDisconnected(handler ServerEventHandler) {
	g.event.register(EVENT_DISCONNECTED, handler)
}
