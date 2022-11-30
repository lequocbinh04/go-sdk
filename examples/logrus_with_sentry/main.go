package main

import (
	goservice "github.com/lequocbinh04/go-sdk"
	"github.com/lequocbinh04/go-sdk/plugin/sckio"
	"log"
)

func main() {
	service := goservice.New(
		goservice.WithName("demo"),
		goservice.WithVersion("1.0.0"),
		goservice.WithSentryDsn(""),
		goservice.WithInitRunnable(sckio.New("sckio")),
	)
	if err := service.Init(); err != nil {
		log.Fatalln(err)
	}

	logger := service.Logger("sckio")
	logger.Errorln("test error1")
	logger.Debugln("test error")

}
