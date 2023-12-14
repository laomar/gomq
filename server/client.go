package server

import (
	"context"
	"fmt"
	"gomq/packets"
	"net"
)

const (
	Connecting = iota
	Connected
)

type client struct {
	ctx           context.Context
	cancel        context.CancelFunc
	conn          net.Conn
	ID            string
	Username      string
	KeepAlive     uint16
	SessionExpiry uint32
	CleanStart    bool
	IP            string
	ConnAt        int64
	Protocol      string
	Version       byte
	Status        byte
	in            chan packets.Packet
	out           chan packets.Packet
}

func (c *client) serve() {
	go c.readLoop()
	go c.writeLoop()
	go c.handleLoop()
	<-c.ctx.Done()
}

func (c *client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	for {
		p, err := packets.ReadPacket(c.conn)
		if err != nil {
			fmt.Println(err)
			break
		}
		c.in <- p
	}
}

func (c *client) writeLoop() {

}

func (c *client) handleLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case in := <-c.in:
			fmt.Println(in)
			switch p := in.(type) {
			case *packets.Connect:
				c.connect(p)
			case *packets.Pingreq:
				c.pingreq()
			}
		}
	}
}

func (c *client) close() {
	c.cancel()
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// Handle connect
func (c *client) connect(pc *packets.Connect) {
	ack := &packets.Connack{
		Version:        c.Version,
		ReasonCode:     packets.Success,
		SessionPresent: pc.CleanStart,
	}
	_ = ack.Pack(c.conn)
}

// Handle ping
func (c *client) pingreq() {
	resp := &packets.Pingresp{}
	_ = resp.Pack(c.conn)
}
