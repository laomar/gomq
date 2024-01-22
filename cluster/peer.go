package cluster

import (
	"context"
	"github.com/hashicorp/serf/serf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"time"
)

type peer struct {
	cluster   *Cluster
	sessionId string
	member    serf.Member
	queue     *queue
	stream    Cluster_SyncClient
	conn      *grpc.ClientConn
	exit      chan bool
}

func (p *peer) stop() {
	select {
	case <-p.exit:
	default:
		close(p.exit)
	}
	_ = p.conn.Close()
}

func (p *peer) start() {
	var err error
	timer := time.NewTimer(0)
	try := 500 * time.Millisecond
	for {
		select {
		case <-p.exit:
			return
		case <-timer.C:
			addr := p.member.Addr.String() + ":" + p.member.Tags["grpc_port"]
			p.conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				timer.Reset(try)
				continue
			}
			cc := NewClusterClient(p.conn)
			if err = p.init(cc); err != nil {
				timer.Reset(try)
				continue
			}
			timer.Stop()
			go p.send()
			go p.recv()
			return
		}
	}
}

func (p *peer) init(cc ClusterClient) error {
	md := metadata.Pairs("NodeName", p.cluster.nodeName)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	rsp, err := cc.Ping(ctx, &PingReq{
		SessionId: p.sessionId,
	})
	if err != nil {
		return err
	}

	if rsp.Restart {
	}

	p.stream, err = cc.Sync(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *peer) send() {
	for {
		select {
		case <-p.exit:
			return
		default:
			es := p.queue.pop(100)
			for _, e := range es {
				err := p.stream.Send(e)
				if err != nil {
					return
				}
			}
		}
	}
}

func (p *peer) recv() {
	for {
		select {
		case <-p.exit:
			return
		default:
			ack, err := p.stream.Recv()
			if err != nil {
				return
			}
			p.queue.del(ack.Id)
		}
	}
}
