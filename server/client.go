package server

import (
	"context"
	"errors"
	"gomq/packets"
	"log"
	"net"
)

const (
	Connecting = iota
	Connected
	Disconnected
)

// Client
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

// Read packet
func (c *client) readPacket() (packets.Packet, error) {
	var fh packets.FixHeader
	if err := fh.Unpack(c.conn); err != nil {
		return nil, err
	}
	p := packets.NewPacket(&fh, c.Version)
	if p == nil {
		return p, errors.New("nil")
	}
	err := p.Unpack(c.conn)
	return p, err
}

func (c *client) readLoop() {
	defer func() {
		//if r := recover(); r != nil {
		//	log.Println("recover", r)
		//}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			p, err := c.readPacket()
			//log.Println(c.ID, p, err)
			if err != nil {
				return
			}
			c.in <- p
		}
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
			switch p := in.(type) {
			case *packets.Connect:
				c.connect(p)
			case *packets.Disconnect:
				c.disconnect(p)
			case *packets.Pingreq:
				c.pingreq()
			case *packets.Publish:
				c.publish(p)
				//default:
				//	log.Println("default handle")
			}
		}
	}
}

// Client close
func (c *client) close() {
	c.cancel()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.Status = Disconnected
}

// Handle connect
func (c *client) connect(pc *packets.Connect) {
	log.Println(pc)
	c.Version = pc.Version
	ta := uint16(8)
	ack := &packets.Connack{
		Version:        pc.Version,
		ReasonCode:     packets.Success,
		SessionPresent: pc.CleanStart,
		Properties:     &packets.Properties{TopicAliasMaximum: &ta},
	}
	_ = ack.Pack(c.conn)
	c.Status = Connected
	c.ID = pc.ClientID
}

// Handle disconnect
func (c *client) disconnect(pd *packets.Disconnect) {
	c.close()
}

// Handle publish
func (c *client) publish(pp *packets.Publish) {
	log.Println(pp)
	ack := &packets.Puback{
		Version:    c.Version,
		ReasonCode: packets.Success,
		PacketID:   pp.PacketID,
	}
	_ = ack.Pack(c.conn)
}

// Handle ping
func (c *client) pingreq() {
	resp := &packets.Pingresp{}
	_ = resp.Pack(c.conn)
}
