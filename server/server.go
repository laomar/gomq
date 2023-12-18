package server

import (
	"context"
	"crypto/tls"
	"golang.org/x/net/websocket"
	. "gomq/config"
	"gomq/log"
	"gomq/packets"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server struct
type Server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	clients *sync.Map
	tcpch   chan bool
	sslch   chan bool
	wsch    chan bool
	wssch   chan bool
}

func New() *Server {
	s := &Server{
		clients: new(sync.Map),
		tcpch:   make(chan bool),
		sslch:   make(chan bool),
		wsch:    make(chan bool),
		wssch:   make(chan bool),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

// Client init
func (s *Server) client(conn net.Conn) *client {
	c := &client{
		server: s,
		conn:   conn,
		IP:     conn.RemoteAddr().String(),
		Status: Connecting,
		in:     make(chan packets.Packet, 8),
		out:    make(chan packets.Packet, 8),
	}
	c.ctx, c.cancel = context.WithCancel(s.ctx)
	return c
}

// TCP server
func (s *Server) tcp() {
	lc, ok := Cfg.Listeners["tcp"]
	if !ok || !lc.Enable {
		return
	}
	ln, err := net.Listen("tcp", lc.Address+":"+lc.Port)
	if err != nil {
		log.Errorf("TCP: %v", err)
		return
	}
	log.Info("TCP: started")
	defer ln.Close()
	for {
		select {
		case <-s.ctx.Done():
			log.Info("TCP: stopped")
			return
		case <-s.tcpch:
			log.Info("TCP: stopped")
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			c := s.client(conn)
			go c.serve()
		}
	}
}

// SSL TCP server
func (s *Server) ssl() {
	lc, ok := Cfg.Listeners["ssl"]
	if !ok || !lc.Enable {
		return
	}
	var cert tls.Certificate
	var err error
	cert, err = tls.LoadX509KeyPair(lc.TLSCert, lc.TLSKey)
	if err != nil {
		log.Errorf("SSL: %v", err)
		return
	}
	var ln net.Listener
	ln, err = tls.Listen("tcp", lc.Address+":"+lc.Port, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Errorf("TCP: %v", err)
		return
	}
	defer ln.Close()
	log.Info("SSL: started")
	for {
		select {
		case <-s.ctx.Done():
			log.Info("SSL: stopped")
			return
		case <-s.sslch:
			log.Info("SSL: stopped")
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			c := s.client(conn)
			go c.serve()
		}
	}
}

// Websocket server
func (s *Server) ws() {
	lc, ok := Cfg.Listeners["ws"]
	if !ok || !lc.Enable {
		return
	}
	router := http.NewServeMux()
	router.Handle(lc.Path, websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		c := s.client(conn)
		c.serve()
	}))
	server := &http.Server{
		Addr:         lc.Address + ":" + lc.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Info("WS: started")
	go func() {
		<-s.wsch
		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorf("WS: %v", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info("WS: stopped")
		} else {
			log.Errorf("WS: %v", err)
		}
	}
}

// Websocket ssl server
func (s *Server) wss() {
	lc, ok := Cfg.Listeners["wss"]
	if !ok || !lc.Enable {
		return
	}
	router := http.NewServeMux()
	router.Handle(lc.Path, websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		c := s.client(conn)
		c.serve()
	}))
	server := &http.Server{
		Addr:         lc.Address + ":" + lc.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Info("WSS: started")
	go func() {
		<-s.wssch
		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorf("WSS: %v", err)
		}
	}()
	if err := server.ListenAndServeTLS(lc.TLSCert, lc.TLSKey); err != nil {
		if err == http.ErrServerClosed {
			log.Info("WSS: stopped")
		} else {
			log.Errorf("WSS: %v", err)
		}
	}
}

// Start server
func (s *Server) Start() {
	log.Info("Server: starting...")
	go s.tcp()
	go s.ssl()
	go s.ws()
	go s.wss()
	<-s.ctx.Done()
	log.Info("Server: stopped")
}

// Restart server
func (s *Server) Restart() {

}

// Stop server
func (s *Server) Stop() {
	s.cancel()
}

// Reload server
func (s *Server) Reload() {
}
