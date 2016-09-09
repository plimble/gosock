package gosock

import (
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kataras/iris"
	"github.com/leavengood/websocket"
	"github.com/raz-varren/sacrificial-socket/log"
	"github.com/valyala/fasthttp"
)

const ( //                        ASCII chars
	startOfHeaderByte uint8 = 1 //SOH
	startOfDataByte         = 2 //STX

	SupportedSubProtocol string = "sac-sock"
)

type Auth func(ctx *iris.Context) error

type event struct {
	eventName    string
	eventHandler func(*Socket, []byte)
}

//SocketServer manages the coordination between
//sockets, rooms, events and the socket hub
type SocketServer struct {
	hub              *socketHub
	events           map[string]*event
	onConnectFunc    func(*Socket)
	onDisconnectFunc func(*Socket)
	l                *sync.RWMutex
}

//NewServer creates a new instance of SocketServer
func NewServer() *SocketServer {
	s := &SocketServer{
		hub:    newHub(),
		events: make(map[string]*event),
		l:      &sync.RWMutex{},
	}

	return s
}

//EnableSignalShutdown listens for linux syscalls SIGHUP, SIGINT, SIGTERM, SIGQUIT, SIGKILL and
//calls the SocketServer.Shutdown() to perform a clean shutdown. true will be passed into complete
//after the Shutdown proccess is finished
func (serv *SocketServer) EnableSignalShutdown(complete chan<- bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL)

	go func() {
		<-c
		complete <- serv.Shutdown()
	}()
}

//Shutdown closes all active sockets and triggers the Shutdown()
//method on any MultihomeBackend that is currently set.
func (serv *SocketServer) Shutdown() bool {
	log.Info.Println("shutting down")
	//complete := serv.hub.shutdown()

	serv.hub.shutdownCh <- true
	socketList := <-serv.hub.socketList

	for _, s := range socketList {
		s.Close()
	}

	if serv.hub.multihomeEnabled {
		log.Info.Println("shutting down multihome backend")
		serv.hub.multihomeBackend.Shutdown()
		log.Info.Println("backend shutdown")
	}

	log.Info.Println("shutdown")
	return true
}

//On registers event functions to be called on individual Socket connections
//when the server's socket receives an Emit from the client's socket.
//
//Any event functions registered with On, must be safe for concurrent use by multiple
//go routines
func (serv *SocketServer) On(eventName string, handleFunc func(*Socket, []byte)) {
	serv.events[eventName] = &event{eventName, handleFunc} //you think you can handle the func?
}

//OnConnect registers an event function to be called whenever a new Socket connection
//is created
func (serv *SocketServer) OnConnect(handleFunc func(*Socket)) {
	serv.onConnectFunc = handleFunc
}

//OnDisconnect registers an event function to be called as soon as a Socket connection
//is closed
func (serv *SocketServer) OnDisconnect(handleFunc func(*Socket)) {
	serv.onDisconnectFunc = handleFunc
}

//WebHandler returns a http.Handler to be passed into http.Handle
func (serv *SocketServer) WebHandler(auth Auth, origins []string) iris.HandlerFunc {
	var upgrader = websocket.FastHTTPUpgrader{
		Handler: serv.loop,
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			if origins == nil || len(origins) == 0 {
				return true
			}

			for _, origin := range origins {
				if origin == string(ctx.Host()) {
					return true
				}
			}

			return false
		},
	}

	return func(ctx *iris.Context) {
		var err error
		if auth != nil {
			if err = auth(ctx); err != nil {
				ctx.Text(401, err.Error())
				return
			}
		}

		upgrader.UpgradeHandler(ctx.RequestCtx)
	}
}

//SetMultihomeBackend registers a MultihomeBackend interface and calls it's Init() method
func (serv *SocketServer) SetMultihomeBackend(b MultihomeBackend) {
	serv.hub.setMultihomeBackend(b)
}

//loop handles all the coordination between new sockets
//reading frames and dispatching events
func (serv *SocketServer) loop(ws *websocket.Conn) {
	s := newSocket(serv, ws)
	log.Info.Println(s.ID(), "connected")
	defer s.Close()

	serv.l.RLock()
	e := serv.onConnectFunc
	serv.l.RUnlock()

	if e != nil {
		e(s)
	}

	for {
		msg, err := s.receive()
		if err == io.EOF {
			return
		}
		if err != nil {
			if err.Error() != "websocket: close 1001 (going away)" {
				log.Err.Println(err)
			}
			return
		}

		eventName := ""
		contentIdx := 0

		for idx, chr := range msg {
			if chr == startOfDataByte {
				eventName = string(msg[:idx])
				contentIdx = idx + 1
				break
			}
		}
		if eventName == "" {
			continue //no event to dispatch
		}

		serv.l.RLock()
		e, exists := serv.events[eventName]
		serv.l.RUnlock()

		if exists {
			go e.eventHandler(s, msg[contentIdx:])
		}
	}
}
