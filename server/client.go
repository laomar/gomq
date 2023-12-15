package server

import (
	"context"
	"errors"
	"gomq/packets"
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
			case *packets.Pubrel:
				c.pubrel(p)
			case *packets.Subscribe:
				c.subscribe(p)
			case *packets.Unsubscribe:
				c.unsubscribe(p)
			case *packets.Auth:
				c.auth(p)
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
	c.Version = pc.Version
	ta := uint16(8)
	ack := &packets.Connack{
		Version:        pc.Version,
		ReasonCode:     packets.Success,
		SessionPresent: false,
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
	switch pp.FixHeader.Qos {
	case packets.Qos0:
	case packets.Qos1:
		ack := &packets.Puback{
			Version:    c.Version,
			ReasonCode: packets.Success,
			PacketID:   pp.PacketID,
		}
		_ = ack.Pack(c.conn)
	case packets.Qos2:
		rec := &packets.Pubrec{
			Version:    c.Version,
			ReasonCode: packets.Success,
			PacketID:   pp.PacketID,
		}
		_ = rec.Pack(c.conn)
	}
}

// Handle pubrel
func (c *client) pubrel(pp *packets.Pubrel) {
	rec := &packets.Pubcomp{
		Version:    c.Version,
		ReasonCode: packets.Success,
		PacketID:   pp.PacketID,
	}
	_ = rec.Pack(c.conn)
}

// Handle Subscribe
func (c *client) subscribe(ps *packets.Subscribe) {
	ack := &packets.Suback{
		Version:  c.Version,
		PacketID: ps.PacketID,
		Payload:  []byte{packets.Success},
	}
	_ = ack.Pack(c.conn)
}

// Handle Unsubscribe
func (c *client) unsubscribe(pu *packets.Unsubscribe) {
	ack := &packets.Unsuback{
		Version:  c.Version,
		PacketID: pu.PacketID,
		Payload:  []byte{packets.Success},
	}
	_ = ack.Pack(c.conn)
}

// Handle auth
func (c *client) auth(pa *packets.Auth) {

}

// Handle ping
func (c *client) pingreq() {
	resp := &packets.Pingresp{}
	_ = resp.Pack(c.conn)
}
