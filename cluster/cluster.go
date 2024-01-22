package cluster

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/logutils"
	"github.com/hashicorp/serf/serf"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/log"
	"github.com/laomar/gomq/pkg/packets"
	"github.com/laomar/gomq/store/topic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cluster struct {
	nodeName    string
	serf        *serf.Serf
	serfEventCh chan serf.Event
	Peers       *sync.Map
	sessions    *sessions
	topicStore  *topicStore
	exit        chan bool
	UnimplementedClusterServer
}

type topicStore struct {
	*topic.Ram
}

func logOut() io.Writer {
	writer := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(config.Cfg.Log.Level)),
		Writer:   log.File(),
	}
	if config.Cfg.Env == "dev" {
		writer.Writer = os.Stdout
	}
	return writer
}

func New() *Cluster {
	return &Cluster{
		nodeName:    config.Cfg.NodeName,
		serfEventCh: make(chan serf.Event, 100),
		Peers:       new(sync.Map),
		sessions: &sessions{
			sessions: make(map[string]*session),
		},
		topicStore: &topicStore{
			topic.NewRam(),
		},
		exit: make(chan bool),
	}
}

func (c *Cluster) Start() error {
	// serf
	logOut := logOut()
	sf := serf.DefaultConfig()
	sf.NodeName = config.Cfg.NodeName
	sf.EventCh = c.serfEventCh
	sf.Tags = map[string]string{"grpc_port": strconv.Itoa(config.Cfg.Cluster.GrpcPort)}
	sf.RejoinAfterLeave = config.Cfg.Cluster.RejoinAfterLeave
	sf.LogOutput = logOut
	sf.MemberlistConfig.BindPort = config.Cfg.Cluster.GossipPort
	sf.MemberlistConfig.AdvertiseAddr = config.Cfg.Cluster.GossipHost
	sf.MemberlistConfig.AdvertisePort = config.Cfg.Cluster.GossipPort
	sf.MemberlistConfig.LogOutput = logOut
	s, err := serf.Create(sf)
	if err != nil {
		return err
	}
	c.serf = s

	// grpc
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(config.Cfg.Cluster.GrpcPort))
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	RegisterClusterServer(srv, c)
	go func() {
		err := srv.Serve(ln)
		if err != nil {
			log.Fatalf("cluster: %s", err)
		}
	}()

	timer := time.NewTimer(0)
	timeout := time.NewTimer(config.Cfg.Cluster.RetryTimeout * time.Second)
	for {
		select {
		case <-timer.C:
			if _, err := c.serf.Join(config.Cfg.Cluster.RetryJoin, true); err != nil {
				timer.Reset(config.Cfg.Cluster.RetryInterval * time.Second)
				continue
			}
			go c.event()
			return nil
		case <-timeout.C:

		}
	}
}

func (c *Cluster) Stop() error {
	if err := c.serf.Leave(); err != nil {
		return err
	}
	if err := c.serf.Shutdown(); err != nil {
		return err
	}
	close(c.exit)
	return nil
}

func (c *Cluster) event() {
	for {
		select {
		case e := <-c.serfEventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				c.join(e.(serf.MemberEvent))
			case serf.EventMemberFailed, serf.EventMemberLeave, serf.EventMemberReap:
				c.leave(e.(serf.MemberEvent))
			case serf.EventMemberUpdate:
			case serf.EventUser:
			case serf.EventQuery:
			default:
			}
		case <-c.exit:
			c.Peers.Range(func(_, v any) bool {
				p := v.(*peer)
				p.stop()
				return true
			})
			return
		}
	}
}

func (c *Cluster) join(me serf.MemberEvent) {
	for _, member := range me.Members {
		if member.Name == c.nodeName {
			continue
		}
		p := &peer{
			cluster:   c,
			sessionId: uuid.NewString(),
			member:    member,
			queue:     newQueue(),
			exit:      make(chan bool),
		}
		if _, ok := c.Peers.LoadOrStore(member.Name, p); !ok {
			go p.start()
			log.Infof("cluster: joined %s %s -> %s", member.Name, member.Addr, config.Cfg.NodeName)
		}
	}
}

func (c *Cluster) leave(me serf.MemberEvent) {
	for _, member := range me.Members {
		if member.Name == c.nodeName {
			continue
		}
		if p, ok := c.Peers.LoadAndDelete(member.Name); ok {
			p.(*peer).stop()
			c.sessions.del(member.Name)
			_ = c.topicStore.UnsubscribeAll(member.Name)
			log.Infof("cluster: left %s %s <- %s", member.Name, member.Addr, config.Cfg.NodeName)
		}
	}
}

func getNodeName(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.DataLoss, "metadata does not exist")
	}
	nodeName := md.Get("NodeName")
	if len(nodeName) > 0 {
		return nodeName[0], nil
	}
	return "", status.Error(codes.DataLoss, "nodename metadata does not exist")
}

func (c *Cluster) Ping(ctx context.Context, req *PingReq) (*PingRsp, error) {
	nodeName, err := getNodeName(ctx)
	if err != nil {
		return nil, err
	}
	if _, ok := c.Peers.Load(nodeName); !ok {
		return nil, status.Errorf(codes.NotFound, "the node %s has not joined", nodeName)
	}

	restart, nextId := c.sessions.set(nodeName, req.SessionId)
	if restart {
	}
	return &PingRsp{
		Restart: restart,
		NextId:  nextId,
	}, nil
}

func (c *Cluster) Sync(stream Cluster_SyncServer) error {
	nodeName, err := getNodeName(stream.Context())
	if err != nil {
		return err
	}
	session := c.sessions.get(nodeName)
	for {
		select {
		case <-session.close:
			return fmt.Errorf("the session of node %s has been closed", nodeName)
		default:
			e, err := stream.Recv()
			if err != nil {
				return err
			}

			c.SyncHandler(session, e)

			err = stream.Send(&Ack{
				Id: e.Id,
			})
			if err != nil {
				return err
			}
			session.nextId = e.Id + 1
		}
	}
}

func (c *Cluster) SyncHandler(s *session, e *Event) {
	// ignore duplicate event
	if s.cache.push(e.Id) {
		return
	}

	// subscribe
	if sub := e.GetSubscribe(); sub != nil {
		topics := strings.Split(sub.Topic, "/")
		shareName := ""
		if len(topics) >= 2 && topics[0] == "$share" {
			shareName = topics[1]
		}
		_, _ = c.topicStore.Subscribe(s.nodeName, &packets.Subscription{
			ShareName: shareName,
			Topic:     sub.Topic,
		})
		//fmt.Println(s.nodeName, shareName, sub.Topic)
		if c.nodeName == "node1" {
			c.topicStore.Print()
		}
	}
}

func (c *Cluster) Subscribe(cid, topic string) {
	c.Peers.Range(func(_, v any) bool {
		p := v.(*peer)
		p.queue.push(&Event{
			Event: &Event_Subscribe{Subscribe: &Subscribe{
				Topic: topic,
			}},
		})
		return true
	})
}
