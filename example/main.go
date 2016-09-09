package main

import (
	"github.com/kataras/iris"
	"github.com/plimble/gosock"
)

func main() {
	s := iris.New()

	ws := gosock.NewServer()

	ws.On("echo", func(ss *gosock.Socket, data []byte) {
		ss.Emit("echo", data)
	})

	s.Get("/socket", ws.WebHandler(nil))

	s.Listen(":9999")
}
