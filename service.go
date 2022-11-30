// Copyright (c) 2019, Viet Tran, 200Lab Team.

package goservice

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/lequocbinh04/go-sdk/httpserver"
	"github.com/lequocbinh04/go-sdk/logger"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	DevEnv     = "dev"
	StgEnv     = "stg"
	PrdEnv     = "prd"
	DefaultEnv = DevEnv
)

type service struct {
	name         string
	version      string
	sentryDsn    string
	env          string
	opts         []Option
	subServices  []Runnable
	initServices map[string]PrefixRunnable
	isRegister   bool
	logger       logger.Logger
	httpServer   HttpServer
	signalChan   chan os.Signal
	cmdLine      *AppFlagSet
	stopFunc     func()
}

func New(opts ...Option) Service {
	sv := &service{
		opts:         opts,
		signalChan:   make(chan os.Signal, 1),
		subServices:  []Runnable{},
		initServices: map[string]PrefixRunnable{},
	}

	for _, opt := range opts {
		opt(sv)
	}

	// init default logger
	if sv.sentryDsn != "" {
		logger.InitServLoggerWithSentryDSN(false, sv.sentryDsn)
	} else {
		logger.InitServLogger(false)
	}
	sv.logger = logger.GetCurrent().GetLogger("service")

	//// Http server
	httpServer := httpserver.New(sv.name, sv.sentryDsn)
	sv.httpServer = httpServer

	sv.subServices = append(sv.subServices, httpServer)

	sv.initFlags()

	if sv.name == "" {
		if len(os.Args) >= 2 {
			sv.name = strings.Join(os.Args[:2], " ")
		}
	}

	loggerRunnable := logger.GetCurrent().(Runnable)
	loggerRunnable.InitFlags()

	sv.cmdLine = newFlagSet(sv.name, flag.CommandLine)
	sv.parseFlags()

	_ = loggerRunnable.Configure()

	return sv
}

func mergeServiceOpts(x []Option, y []Option) []Option {
	z := make([]Option, len(x)+len(y))
	copy(z, x)
	copy(z[len(x):], y)
	return z
}

func (sv *service) Add(opts ...Option) Service {
	for _, opt := range opts {
		opt(sv)
	}
	sv.opts = mergeServiceOpts(sv.opts, opts)
	return sv
}

func (sv *service) Name() string {
	return sv.name
}

func (sv *service) Version() string {
	return sv.version
}

func (sv *service) Init() error {
	for _, dbSv := range sv.initServices {
		if err := dbSv.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (sv *service) InitPrefix(prefix ...string) error {
	for _, pre := range prefix {
		if err := sv.initServices[pre].Run(); err != nil {
			return err
		}
	}
	return nil
}

func (sv *service) IsRegistered() bool {
	return sv.isRegister
}

func (sv *service) Start() error {
	signal.Notify(sv.signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	c := sv.run()
	//s.stopFunc = s.activeRegistry()

	for {
		select {
		case err := <-c:
			if err != nil {
				sv.logger.Error(err.Error())
				sv.Stop()
				return err
			}

		case sig := <-sv.signalChan:
			sv.logger.Infoln(sig)
			switch sig {
			case syscall.SIGHUP:
				return nil
			default:
				sv.Stop()
				return nil
			}
		}
	}
}

func (sv *service) initFlags() {
	flag.StringVar(&sv.env, "app-env", DevEnv, "Env for service. Ex: dev | stg | prd")

	for _, subService := range sv.subServices {
		subService.InitFlags()
	}

	for _, dbService := range sv.initServices {
		dbService.InitFlags()
	}
}

// Run service and its components at the same time
func (sv *service) run() <-chan error {
	c := make(chan error, 1)

	// Start all services
	for _, subService := range sv.subServices {
		go func(subSv Runnable) { c <- subSv.Run() }(subService)
	}

	return c
}

// Stop service and stop its components at the same time
func (sv *service) Stop() {
	sv.logger.Infoln("Stopping service...")
	stopChan := make(chan bool)
	for _, subService := range sv.subServices {
		go func(subSv Runnable) { stopChan <- <-subSv.Stop() }(subService)
	}

	for _, dbSv := range sv.initServices {
		go func(subSv Runnable) { stopChan <- <-subSv.Stop() }(dbSv)
	}

	for i := 0; i < len(sv.subServices)+len(sv.initServices); i++ {
		<-stopChan
	}

	//s.stopFunc()
	sv.logger.Infoln("service stopped")
}

func (sv *service) RunFunction(fn Function) error {
	return fn(sv)
}

func (sv *service) HTTPServer() HttpServer {
	return sv.httpServer
}

func (sv *service) Logger(prefix string) logger.Logger {
	return logger.GetCurrent().GetLogger(prefix)
}

func (sv *service) OutEnv() {
	sv.cmdLine.GetSampleEnvs()
}

func (sv *service) parseFlags() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	_, err := os.Stat(envFile)
	if err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			sv.logger.Fatalf("Loading env(%s): %s", envFile, err.Error())
		}
	} else if envFile != ".env" {
		sv.logger.Fatalf("Loading env(%s): %s", envFile, err.Error())
	}

	sv.cmdLine.Parse([]string{})
}

// WithName Service must have a name for service discovery and logging/monitoring
func WithName(name string) Option {
	return func(s *service) { s.name = name }
}

// WithVersion Every deployment needs a specific version
func WithVersion(version string) Option {
	return func(s *service) { s.version = version }
}

func WithSentryDsn(dsn string) Option {
	return func(s *service) { s.sentryDsn = dsn }
}

// Service will write log data to file with this option
func WithFileLogger() Option {
	return func(s *service) {
		logger.InitServLogger(true)
	}
}

// Add Runnable component to SDK
// These components will run parallel in when service run
func WithRunnable(r Runnable) Option {
	return func(s *service) { s.subServices = append(s.subServices, r) }
}

// Add init component to SDK
// These components will run sequentially before service run
func WithInitRunnable(r PrefixRunnable) Option {
	return func(s *service) {
		if _, ok := s.initServices[r.GetPrefix()]; ok {
			log.Fatal(fmt.Sprintf("prefix %s is duplicated", r.GetPrefix()))
		}

		s.initServices[r.GetPrefix()] = r
	}
}

func (sv *service) Get(prefix string) (interface{}, bool) {
	is, ok := sv.initServices[prefix]

	if !ok {
		return nil, ok
	}

	return is.Get(), true
}

func (sv *service) MustGet(prefix string) interface{} {
	db, ok := sv.Get(prefix)

	if !ok {
		panic(fmt.Sprintf("can not get %s\n", prefix))
	}

	return db
}

func (sv *service) Env() string { return sv.env }
