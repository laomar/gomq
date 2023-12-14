package server

import (
	"context"
	"gomq/packets"
	"net"
	"sync"
)

type server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	clients *sync.Map
	ch      chan int
}

func New() *server {
	s := &server{}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *server) client(conn net.Conn) *client {
	c := &client{
		conn:   conn,
		IP:     conn.RemoteAddr().String(),
		Status: Connecting,
		in:     make(chan packets.Packet, 8),
		out:    make(chan packets.Packet, 8),
	}
	c.ctx, c.cancel = context.WithCancel(s.ctx)
	return c
}

func (s *server) tcp() {
	l, _ := net.Listen("tcp", ":1883")
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		c := s.client(conn)
		go c.serve()
	}
}

func (s *server) websocket() {
}

func (s *server) Start() {
	go s.tcp()
	go s.websocket()
	<-s.ctx.Done()
}

func (s *server) Stop() {
}

func (s *server) Reload() {
}
