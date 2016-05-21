package gosocklite

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type GoSock struct {
	hub   *hub
	event *eventEmitter
}

func New() *GoSock {
	return &GoSock{}
}

func (g *GoSock) Handler() http.HandlerFunc {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	go runHub()
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		c := newConnection(ws)
		h.register <- c
		evt.emit(EVENT_CONNECTION, nil, c)
		go c.writePump()
		c.readPump()
	}
}

func (g *GoSock) On(event string, handler EventHandler) {
	evt.register(event, handler)
}

func (g *GoSock) OnConnection(handler EventHandler) {
	evt.register(EVENT_CONNECTION, handler)
}

func (g *GoSock) OnDisconnected(handler EventHandler) {
	evt.register(EVENT_DISCONNECTED, handler)
}
