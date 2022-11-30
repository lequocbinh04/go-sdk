package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/plugin/sckio"
	"log"
)

func main() {
	service := goservice.New(
		goservice.WithName("demo"),
		goservice.WithVersion("1.0.0"),
		goservice.WithSentryDsn("https://7c4fda062eac4dbca558c2e262ad649b@o4504246471950336.ingest.sentry.io/4504246473457664"),
		goservice.WithInitRunnable(sckio.New("sckio")),
	)
	if err := service.Init(); err != nil {
		log.Fatalln(err)
	}

	//logger := service.Logger("sckio")
	//logger.Errorln("test error1")
	//logger.Debugln("test error")

	service.HTTPServer().AddHandler(func(r *gin.Engine) {
		r.Use(func(c *gin.Context) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("panic", err)
					panic(err) // panic for sentry
				}
			}()
			c.Next()
		})
		r.GET("/ping", func(c *gin.Context) {
			panic(errors.New("err db1"))
		})
	})

	service.Start()

}
