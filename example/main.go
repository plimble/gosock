package main

import (
	"log"
	"net/http"

	"github.com/kr/pretty"
	"github.com/plimble/gosock"
)

func main() {
	g := gosock.New()
	g.OnConnection(func(data []byte, c *gosock.Connection) {
		pretty.Println("connected!!!")
	})

	g.OnError(func(err error, c *gosock.Connection) {
		pretty.Println(err)
	})

	g.OnDisconnected(func(data []byte, c *gosock.Connection) {
		pretty.Println("disconnected!!!")
	})

	g.On("broadcast", func(data []byte, c *gosock.Connection) {
		c.Broadcast("getbroadcast", data)
	})

	g.On("broadcastroom", func(data []byte, c *gosock.Connection) {
		c.BroadcastRoom("rooma", "broadcastroom", data)
	})

	g.On("pm", func(data []byte, c *gosock.Connection) {
		c.To("1", "getpm", data)
	})

	g.On("req", func(data []byte, c *gosock.Connection) {
		c.Reply("getreq", data)
	})

	http.HandleFunc("/", g.Handler())
	log.Fatal(http.ListenAndServe(":3000", nil))
}
