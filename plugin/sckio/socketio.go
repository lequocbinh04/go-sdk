package sckio

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/logger"
	"log"
	"sync"
)

type SocketServer interface {
	UserSockets(userId int64) []AppSocket
	EmitToRoom(room string, key string, data interface{}) error
	EmitToUser(userId int64, key string, data interface{}) error
	StartRealtimeServer(engine *gin.Engine, sc goservice.ServiceContext, op ObserverProvider)
	GetSocketServer() *socketio.Server
	SaveAppSocket(userId int64, appSck AppSocket)
}

type Config struct {
	Name          string
	MaxConnection int
}

type sckServer struct {
	Config
	io      *socketio.Server
	logger  logger.Logger
	storage map[int64][]AppSocket
	locker  *sync.RWMutex
}

func New(name string) *sckServer {
	return &sckServer{
		Config:  Config{Name: name},
		storage: make(map[int64][]AppSocket),
		locker:  new(sync.RWMutex),
	}
}

type ObserverProvider interface {
	AddObservers(server *socketio.Server, sc goservice.ServiceContext, l logger.Logger) func(socketio.Conn) error
}

func (s *sckServer) StartRealtimeServer(engine *gin.Engine, sc goservice.ServiceContext, op ObserverProvider) {
	server, err := socketio.NewServer(nil)

	if err != nil {
		s.logger.Fatal(err)
	}

	s.io = server
	server.OnConnect("/", op.AddObservers(server, sc, s.logger))

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go server.Serve()

	engine.GET("/socket.io/*any", gin.WrapH(server))
	engine.POST("/socket.io/*any", gin.WrapH(server))
	engine.Handle("WS", "/socket.io/*any", gin.WrapH(server))
	engine.Handle("WSS", "/socket.io/*any", gin.WrapH(server))
}

func (s *sckServer) UserSockets(userId int64) []AppSocket {
	var sockets []AppSocket

	if scks, ok := s.storage[userId]; ok {
		return scks
	}

	return sockets
}

func (s *sckServer) EmitToRoom(room string, key string, data interface{}) error {
	s.io.BroadcastToRoom("/", room, key, data)
	return nil
}

func (s *sckServer) SaveAppSocket(userId int64, appSck AppSocket) {
	s.locker.Lock()

	//appSck.Join("order-{ordID}")

	if v, ok := s.storage[userId]; ok {
		s.storage[userId] = append(v, appSck)
	} else {
		s.storage[userId] = []AppSocket{appSck}
	}

	s.locker.Unlock()
}

func (s *sckServer) getAppSocket(userId int64) []AppSocket {
	s.locker.RLock()
	defer s.locker.RUnlock()

	return s.storage[userId]
}

func (s *sckServer) EmitToUser(userId int64, key string, data interface{}) error {
	sockets := s.getAppSocket(userId)

	for _, s := range sockets {
		s.Emit(key, data)
	}

	return nil
}

func (s *sckServer) GetSocketServer() *socketio.Server {
	return s.io
}

func (s *sckServer) GetPrefix() string {
	return s.Config.Name
}

func (s *sckServer) Get() interface{} {
	return s
}

func (s *sckServer) Name() string {
	return s.Config.Name
}

func (s *sckServer) InitFlags() {
	pre := s.GetPrefix()
	flag.IntVar(&s.MaxConnection, fmt.Sprintf("%s-max-connection", pre), 2000, "socket max connection")
}

func (s *sckServer) Configure() error {
	s.logger = logger.GetCurrent().GetLogger("io.socket")
	return nil
}

func (s *sckServer) Run() error {
	return s.Configure()
}

func (s *sckServer) Stop() <-chan bool {
	c := make(chan bool)
	go func() { c <- true }()
	return c
}
