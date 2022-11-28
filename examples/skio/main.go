package main

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/logger"
	"github.com/lequocbinh04/go-sdk/plugin/sckio"
	"log"
)

type observerProvider struct{}

func (o *observerProvider) AddObservers(server *socketio.Server, sc goservice.ServiceContext, l logger.Logger) func(socketio.Conn) error {
	return func(conn socketio.Conn) error {
		l.Infoln("New connection", conn.ID(), ", IP:", conn.RemoteAddr())
		return nil
	}
}

func main() {
	service := goservice.New(
		goservice.WithName("demo"),
		goservice.WithVersion("1.0.0"),
		goservice.WithInitRunnable(sckio.New("sckio")),
	)
	if err := service.Init(); err != nil {
		log.Fatalln(err)
	}

	sckio := service.MustGet("sckio").(sckio.SocketServer)
	service.HTTPServer().AddHandler(func(r *gin.Engine) {
		var observer *observerProvider
		sckio.StartRealtimeServer(r, service, observer)
		r.StaticFile("/demo/", "./examples/skio/demo.html")
		r.GET("/ping", func(c *gin.Context) {
			c.String(200, "pong")
		})
	})

	if err := service.Start(); err != nil {
		log.Fatalln(err)
	}
}

//func main() {
//	r := gin.Default()
//	r.StaticFile("/demo/", "./examples/skio/demo.html")
//	r.GET("/ping", func(c *gin.Context) {
//		c.String(200, "pong")
//	})
//	rtEngine := skio.NewEngine()
//	if err := rtEngine.Run(r); err != nil {
//		log.Fatalln(err)
//	}
//	r.Run(":3000")
//}
