package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
	. "gomq/config"
	"gomq/log"
	"gomq/packets"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Server struct
type Server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	gw      sync.WaitGroup
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
		s.gw.Done()
		return
	}
	ln, err := net.Listen("tcp", lc.Address+":"+lc.Port)
	if err != nil {
		log.Errorf("TCP: %v", err)
		return
	}
	defer ln.Close()
	log.Info("TCP: started")
	s.gw.Done()
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
		s.gw.Done()
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
	s.gw.Done()
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
		s.gw.Done()
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
	go func() {
		<-s.wsch
		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorf("WS: %v", err)
		}
	}()
	log.Info("WS: started")
	s.gw.Done()
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
		s.gw.Done()
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
	go func() {
		<-s.wssch
		if err := server.Shutdown(s.ctx); err != nil {
			log.Errorf("WSS: %v", err)
		}
	}()
	log.Info("WSS: started")
	s.gw.Done()
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
	s.gw.Add(5)
	go s.tcp()
	go s.ssl()
	go s.ws()
	go s.wss()
	go s.signal()
	s.gw.Wait()
	log.Info("Server: started")
	<-s.ctx.Done()
	log.Info("Server: stopped")
}

// Stop server
func (s *Server) Stop() {
	s.cancel()
	os.Remove(Cfg.PidFile)
}

// Reload server
func (s *Server) Reload() {
	ParseConfig()
	log.Info("Server: reloaded")
}

// Signal process
func (s *Server) signal() {
	s.gw.Done()
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	reload := make(chan os.Signal)
	signal.Notify(reload, syscall.SIGHUP)
	for {
		select {
		case <-stop:
			s.Stop()
		case <-reload:
			s.Reload()
		}
	}
}

// ReloadCmd create reload command
func ReloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reload",
		Short: "Reload gomq broker",
		Run: func(cmd *cobra.Command, args []string) {
			bs, err := os.ReadFile(Cfg.PidFile)
			if err != nil {
				log.Errorf("Server reload: %v", err)
				return
			}
			pid, err := strconv.Atoi(string(bs))
			if err != nil {
				log.Errorf("Server reload: %v", err)
				return
			}
			p, err := os.FindProcess(pid)
			if err != nil {
				log.Errorf("Server reload: %v", err)
				return
			}
			if err := p.Signal(syscall.SIGHUP); err != nil {
				log.Errorf("Server reload: %v", err)
			}
		},
	}
}

// StopCmd create stop command
func StopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop gomq broker",
		Run: func(cmd *cobra.Command, args []string) {
			bs, err := os.ReadFile(Cfg.PidFile)
			if err != nil {
				log.Errorf("Server stop: %v", err)
				return
			}
			pid, err := strconv.Atoi(string(bs))
			if err != nil {
				log.Errorf("Server stop: %v", err)
				return
			}
			p, err := os.FindProcess(pid)
			if err != nil {
				log.Errorf("Server stop: %v", err)
				return
			}
			if err := p.Signal(os.Interrupt); err != nil {
				log.Errorf("Server stop: %v", err)
			}
		},
	}
}

// StartCmd create start command
func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start gomq broker",
		Run: func(cmd *cobra.Command, args []string) {
			piddir := filepath.Dir(Cfg.PidFile)
			if _, err := os.Stat(piddir); os.IsNotExist(err) {
				if err := os.MkdirAll(piddir, 0755); err != nil {
					log.Fatalf("Server start: %v", err)
				}
			}
			pid := fmt.Sprintf("%d", os.Getpid())
			if err := os.WriteFile(Cfg.PidFile, []byte(pid), 0644); err != nil {
				log.Fatalf("Server start: %v", err)
			}
			server := New()
			server.Start()
		},
	}
}
