package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/log"
	"github.com/laomar/gomq/pkg/packets"
	"github.com/laomar/gomq/store"
	"github.com/laomar/gomq/store/topic"
	"github.com/pires/go-proxyproto"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var plugins = make(map[string]Plugin)

// Server struct
type Server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	clients *sync.Map

	topicStore topic.Store

	cancelTcp context.CancelFunc
	cancelSsl context.CancelFunc
	cancelWs  context.CancelFunc
	cancelWss context.CancelFunc
}

func New() *Server {
	s := &Server{
		clients: new(sync.Map),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func RegPlugin(name string, new NewPlugin) {
	if _, ok := plugins[name]; !ok {
		plugin, err := new()
		if err != nil {
			log.Errorf("plugin: register %s", name)
		}
		plugins[name] = plugin
	}
}

func (s *Server) loadPlugin() error {
	for _, plugin := range plugins {
		log.Infof("plugin: loading %s", plugin.Name())
		if err := plugin.Load(); err != nil {
			return err
		}
	}
	return nil
}

// Init Server
func (s *Server) Init() error {
	var err error

	var se store.Store
	if se, err = store.NewStore(); err != nil {
		log.Fatalf("store: %v", err)
	}
	if s.topicStore, err = se.NewTopicStore(); err != nil {
		log.Fatalf("topic store: %v", err)
	}
	s.topicStore.Init("mqttx_bf87b741")

	return s.loadPlugin()
}

// New Client
func (s *Server) NewClient(ctx context.Context, conn net.Conn) *client {
	c := &client{
		server: s,
		conn:   conn,
		Status: Connecting,
		in:     make(chan packets.Packet, 16),
		out:    make(chan packets.Packet, 16),
		prop:   new(clientProp),
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	return c
}

// TCP server
func (s *Server) tcp() {
	lc, ok := config.Cfg.Listeners["tcp"]
	if !ok || !lc.Enable {
		return
	}
	ln, err := net.Listen("tcp", lc.Addr)
	if err != nil {
		log.Errorf("tcp: %v", err)
		return
	}
	var pln net.Listener
	if lc.ProxyProtocol {
		pln = &proxyproto.Listener{Listener: ln}
	} else {
		pln = ln
	}
	defer pln.Close()
	var ctx context.Context
	ctx, s.cancelTcp = context.WithCancel(s.ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := pln.Accept()
				if err != nil {
					continue
				}
				c := s.NewClient(ctx, conn)
				go c.serve()
			}
		}
	}()
	log.Infof("tcp: listening [%s]", lc.Addr)
	<-ctx.Done()
	log.Info("tcp: closed")
}

// TLS TCP server
func (s *Server) tls() {
	lc, ok := config.Cfg.Listeners["tls"]
	if !ok || !lc.Enable {
		return
	}
	var cert tls.Certificate
	var err error
	cert, err = tls.LoadX509KeyPair(lc.TLSCert, lc.TLSKey)
	if err != nil {
		log.Errorf("tls: %v", err)
		return
	}
	var ln net.Listener
	ln, err = tls.Listen("tcp", lc.Addr, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Errorf("tls: %v", err)
		return
	}
	var pln net.Listener
	if lc.ProxyProtocol {
		pln = &proxyproto.Listener{Listener: ln}
	} else {
		pln = ln
	}
	defer pln.Close()
	var ctx context.Context
	ctx, s.cancelSsl = context.WithCancel(s.ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := pln.Accept()
				if err != nil {
					continue
				}
				c := s.NewClient(ctx, conn)
				go c.serve()
			}
		}
	}()
	log.Infof("tls: listening [%s]", lc.Addr)
	<-ctx.Done()
	log.Info("tls: closed")
}

// Websocket server
func (s *Server) ws() {
	lc, ok := config.Cfg.Listeners["ws"]
	if !ok || !lc.Enable {
		return
	}
	var ctx context.Context
	ctx, s.cancelWs = context.WithCancel(s.ctx)
	router := http.NewServeMux()
	router.Handle(lc.Path, websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		c := s.NewClient(ctx, conn)
		c.serve()
	}))
	server := &http.Server{
		Addr:         lc.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Errorf("ws: %v", err)
		return
	}
	var pln net.Listener
	if lc.ProxyProtocol {
		pln = &proxyproto.Listener{
			Listener:          ln,
			ReadHeaderTimeout: 10 * time.Second,
		}
	} else {
		pln = ln
	}
	defer pln.Close()
	go func() {
		if err := server.Serve(pln); err != nil && err != http.ErrServerClosed {
			log.Errorf("ws: %v", err)
		}
	}()
	log.Infof("ws: listening [%s]", lc.Addr)
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("ws: %v", err)
	}
	log.Info("ws: closed")
}

// Websocket ssl server
func (s *Server) wss() {
	lc, ok := config.Cfg.Listeners["wss"]
	if !ok || !lc.Enable {
		return
	}
	var ctx context.Context
	ctx, s.cancelWss = context.WithCancel(s.ctx)
	router := http.NewServeMux()
	router.Handle(lc.Path, websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		c := s.NewClient(ctx, conn)
		c.serve()
	}))
	server := &http.Server{
		Addr:         lc.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Errorf("wss: %v", err)
		return
	}
	var pln net.Listener
	if lc.ProxyProtocol {
		pln = &proxyproto.Listener{
			Listener:          ln,
			ReadHeaderTimeout: 10 * time.Second,
		}
	} else {
		pln = ln
	}
	defer pln.Close()
	go func() {
		if err := server.ServeTLS(pln, lc.TLSCert, lc.TLSKey); err != nil && err != http.ErrServerClosed {
			log.Errorf("wss: %v", err)
		}
	}()
	log.Infof("wss: listening [%s]", lc.Addr)
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("wss: %v", err)
	}
	log.Info("wss: closed")
}

// Start server
func (s *Server) Start() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	reload := make(chan os.Signal)
	signal.Notify(reload, syscall.SIGHUP)

	log.Info("gomq: starting...")
	go s.tcp()
	go s.tls()
	go s.ws()
	go s.wss()
	go s.Pprof()

	for {
		select {
		case <-s.ctx.Done():
			time.Sleep(time.Second)
			log.Info("gomq: stopped")
			return
		case <-stop:
			os.Remove(config.Cfg.PidFile)
			s.Stop()
		case <-reload:
			s.Reload()
		}
	}
}

// Stop server
func (s *Server) Stop() {
	defer s.cancel()
}

// Reload server
func (s *Server) Reload() {
	log.Info("gomq: reload...")
	config.Cfg.Parse()
	//log.Init()

	s.cancelTcp()
	s.cancelSsl()
	s.cancelWs()
	s.cancelWss()
	time.Sleep(1 * time.Second)
	go s.tcp()
	go s.tls()
	go s.ws()
	go s.wss()
}

// Pprof Listen
func (s *Server) Pprof() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
			log.Errorf("pprof: %v", err)
		}
	}()
	log.Info("pprof: listening")
	<-s.ctx.Done()
	log.Info("pprof: closed")
}

// ReloadCmd create reload command
func ReloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reload",
		Short: "Reload gomq broker",
		Run: func(cmd *cobra.Command, args []string) {
			bs, err := os.ReadFile(config.Cfg.PidFile)
			if err != nil {
				log.Errorf("gomq reload: %v", err)
				return
			}
			pid, err := strconv.Atoi(string(bs))
			if err != nil {
				log.Errorf("gomq reload: %v", err)
				return
			}
			p, err := os.FindProcess(pid)
			if err != nil {
				log.Errorf("gomq reload: %v", err)
				return
			}
			if err := p.Signal(syscall.SIGHUP); err != nil {
				log.Errorf("gomq reload: %v", err)
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
			bs, err := os.ReadFile(config.Cfg.PidFile)
			if err != nil {
				log.Errorf("gomq stop: %v", err)
				return
			}
			pid, err := strconv.Atoi(string(bs))
			if err != nil {
				log.Errorf("gomq stop: %v", err)
				return
			}
			p, err := os.FindProcess(pid)
			if err != nil {
				log.Errorf("gomq stop: %v", err)
				return
			}
			if err := p.Signal(syscall.SIGINT); err != nil {
				log.Errorf("gomq stop: %v", err)
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
			piddir := filepath.Dir(config.Cfg.PidFile)
			if _, err := os.Stat(piddir); os.IsNotExist(err) {
				if err := os.MkdirAll(piddir, 0755); err != nil {
					log.Fatalf("gomq start: %v", err)
				}
			}
			pid := fmt.Sprintf("%d", os.Getpid())
			if err := os.WriteFile(config.Cfg.PidFile, []byte(pid), 0644); err != nil {
				log.Fatalf("gomq start: %v", err)
			}
			server := New()
			if err := server.Init(); err != nil {
				log.Fatalf("gomq start: %v", err)
			}
			server.Start()
		},
	}
}
