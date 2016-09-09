package gosockredis

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/plimble/gosock"
	"github.com/raz-varren/sacrificial-socket/log"
)

type Redis struct {
	psc redis.PubSubConn
	c   redis.Conn
	ch  string
	bc  string
	rc  string
}

func NewRedis(ch string, p *redis.Pool) *Redis {
	return &Redis{redis.PubSubConn{Conn: p.Get()}, p.Get(), ch, ch + "." + "bc", ch + "." + "rc"}
}

func (r *Redis) Init() {
	r.psc.Subscribe(r.bc)
	r.psc.Subscribe(r.rc)
}

func (r *Redis) Shutdown() {
	r.psc.Unsubscribe(r.bc)
	r.psc.Unsubscribe(r.rc)
}

//BroadcastToBackend is called everytime a BroadcastMsg is
//sent by a Socket
//
//BroadcastToBackend must be safe for concurrent use by multiple
//go routines
func (r *Redis) BroadcastToBackend(b *gosock.BroadcastMsg) {
	msg, _ := json.Marshal(&BCData{
		Event: b.EventName,
		Data:  b.Data,
	})

	r.c.Send("PUBLISH", r.bc, msg)
	r.c.Flush()
}

//RoomcastToBackend is called everytime a RoomMsg is sent
//by a socket, even if none of this server's sockets are
//members of that room
//
//RoomcastToBackend must be safe for concurrent use by multiple
//go routines
func (r *Redis) RoomcastToBackend(rr *gosock.RoomMsg) {
	msg, _ := json.Marshal(&RCData{
		Room:  rr.RoomName,
		Event: rr.EventName,
		Data:  rr.Data,
	})

	r.c.Send("PUBLISH", r.rc, msg)
	r.c.Flush()
}

type BCData struct {
	Event string      `json:"e"`
	Data  interface{} `json:"d"`
}

//BroadcastFromBackend is called once and only once as a go routine as
//soon as the MultihomeBackend is registered using
//SocketServer.SetMultihomeBackend
//
//b consumes a BroadcastMsg and dispatches
//it to all sockets on this server
func (r *Redis) BroadcastFromBackend(b chan<- *gosock.BroadcastMsg) {
	for {
		switch n := r.psc.Receive().(type) {
		case redis.Message:
			log.Info.Printf("Message: %s %s\n", n.Channel, n.Data)

			bcast := &BCData{}
			json.Unmarshal(n.Data, bcast)

			switch n.Channel {
			case r.bc:
				b <- &gosock.BroadcastMsg{EventName: bcast.Event, Data: bcast.Data}
			}
		case redis.Subscription:
			log.Info.Printf("%s: %s %d\n", n.Channel, n.Kind, n.Count)
		case error:
			log.Err.Printf("error: %v\n", n)
			return
		}
	}
}

type RCData struct {
	Room  string      `json:"r"`
	Event string      `json:"e"`
	Data  interface{} `json:"d"`
}

//RoomcastFromBackend is called once and only once as a go routine as
//soon as the MultihomeBackend is registered using
//SocketServer.SetMultihomeBackend
//
//r consumes a RoomMsg and dispatches it to all sockets
//that are members the specified room
func (r *Redis) RoomcastFromBackend(rr chan<- *gosock.RoomMsg) {
	for {
		switch n := r.psc.Receive().(type) {
		case redis.Message:
			log.Info.Printf("Message: %s %s\n", n.Channel, n.Data)

			rcast := &RCData{}
			json.Unmarshal(n.Data, rcast)

			switch n.Channel {
			case r.rc:
				rr <- &gosock.RoomMsg{RoomName: rcast.Room, EventName: rcast.Event, Data: rcast.Data}
			}
		case redis.Subscription:
			log.Info.Printf("%s: %s %d\n", n.Channel, n.Kind, n.Count)
		case error:
			log.Err.Printf("error: %v\n", n)
			return
		}
	}
}
