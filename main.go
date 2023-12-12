package main

import "fmt"

type Server interface {
	wa()
}

type server struct {
	a string
	b int
}

func (s server) wa(a string) {
	s.a = a
}

type options func(srv *server)

func witha(a string) options {
	return func(srv *server) {
		srv.a = a
	}
}

func new(opts ...options) *server {
	srv := &server{b: 5}
	for _, fn := range opts {
		fn(srv)
	}
	return srv
}

func main() {
	srv := new(witha("test"))
	fmt.Println(srv)
}
