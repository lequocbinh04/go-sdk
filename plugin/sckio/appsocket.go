package sckio

import (
	"github.com/lequocbinh04/go-sdk/sdkcm"
	"net"
	"net/http"
	"net/url"
)

type Socket interface {
	ID() string
	Close() error
	URL() url.URL
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	RemoteHeader() http.Header

	Context() interface{}
	SetContext(v interface{})
	Namespace() string
	Emit(msg string, v ...interface{})

	Join(room string)
	Leave(room string)
	LeaveAll()
	Rooms() []string
}

type AppSocket interface {
	CurrentUser() sdkcm.Requester
	Socket
}

type appSocket struct {
	Socket
}

func NewAppSocket(s Socket, user sdkcm.Requester) *appSocket {
	s.SetContext(user)
	return &appSocket{s}
}

func (s *appSocket) CurrentUser() sdkcm.Requester {
	return s.Context().(sdkcm.Requester)
}
