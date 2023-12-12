package server

import (
	"net"
	"sync"
)

type server struct {
	wg sync.WaitGroup
}

func New() *server {
	s := &server{}
	return s
}

func (s *server) TCP() {
	ln, _ := net.Listen("tcp", ":1883")
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		client := NewClient(conn)
		go client.ReadLoop()
	}
}

func (s *server) Websocket() {

}

func (s *server) Run() {
	s.wg.Add(2)
	go s.TCP()
	go s.Websocket()
	s.wg.Wait()
}

func (s *server) Stop() {

}
