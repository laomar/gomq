package server

import (
	"context"
	"fmt"
	. "gomq/config"
	"gomq/log"
	"gomq/packets"
	"math"
	"net"
	"time"
)

const (
	Connecting = iota
	Connected
	Disconnected
)

// Client
type clientProp struct {
	KeepAlive             uint16
	SessionExpiryInterval uint32
	MaxInflight           uint16
	MaximumPacketSize     uint32
	TopicAliasMaximum     uint16
	Username              string
	Protocol              string
	CleanStart            bool
	IP                    string
}
type client struct {
	ctx     context.Context
	cancel  context.CancelFunc
	server  *Server
	conn    net.Conn
	ID      string
	ConnAt  int64
	Version byte
	Status  byte
	in      chan packets.Packet
	out     chan packets.Packet
	prop    *clientProp
}

func (c *client) serve() {
	go c.readLoop()
	go c.writeLoop()
	if c.connect() {
		go c.handleLoop()
	}
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
		return p, packets.ErrProtocol
	}
	err := p.Unpack(c.conn)
	return p, err
}

// Write packet
func (c *client) writePacket(p packets.Packet) error {
	return p.Pack(c.conn)
}

func (c *client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Client read recover: %v", r)
		}
		close(c.in)
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if ka := c.prop.KeepAlive; ka > 0 {
				_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(ka/2+ka)))
			}
			p, err := c.readPacket()
			if err != nil {
				log.Debugf("Client read: %v", err)
				c.close()
				return
			}
			c.in <- p
		}
	}
}

func (c *client) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Client write recover: %v", r)
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case out := <-c.out:
			err := c.writePacket(out)
			if err != nil {
				log.Debugf("Client write: %v", err)
				c.close()
				return
			}
		}
	}
}

func (c *client) handleLoop() {
	for in := range c.in {
		switch p := in.(type) {
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

// Client close
func (c *client) close() {
	defer c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
	count := 0
	c.server.clients.Range(func(id, c any) bool {
		count++
		return true
	})
	//fmt.Println(count)
	c.server.clients.Delete(c.ID)
}

// Handle connect
func (c *client) connect() bool {
	var pc *packets.Connect
	for in := range c.in {
		var code byte
		switch p := in.(type) {
		case *packets.Connect:
			pc = p
			if len(pc.ClientID) == 0 {
				code = packets.ClientIdentifierNotValid
				break
			}
			code = c.connectHandler(pc)
		case *packets.Auth:
		default:
			code = packets.MalformedPacket
		}

		c.Version = pc.Version

		// continue authentication
		if code == packets.ContinueAuthentication {
			auth := &packets.Auth{
				ReasonCode: code,
				Properties: &packets.Properties{
					AuthMethod: pc.Properties.AuthMethod,
				},
			}
			c.out <- auth
			continue
		}

		// connect fail
		ack := &packets.Connack{
			Version:    pc.Version,
			ReasonCode: code,
		}
		if code != packets.Success {
			c.out <- ack
			return false
		}

		// connect success
		c.Status = Connected
		c.ID = pc.ClientID
		c.prop.Username = pc.Username
		c.prop.Protocol = pc.Protocol
		c.prop.CleanStart = pc.CleanStart
		c.prop.IP = c.conn.RemoteAddr().String()
		c.ConnAt = time.Now().Unix()
		c.prop.MaxInflight = Cfg.Mqtt.MaxInflight

		if c.Version == packets.V5 {
			c.prop.KeepAlive = Cfg.Mqtt.ServerKeepAlive
			if pc.KeepAlive < c.prop.KeepAlive || c.prop.KeepAlive == 0 {
				c.prop.KeepAlive = pc.KeepAlive
			}

			if rm := pc.Properties.ReceiveMaximum; rm != nil {
				c.prop.MaxInflight = *rm
			}

			c.prop.SessionExpiryInterval = Cfg.Mqtt.SessionExpiryInterval
			if sei := pc.Properties.SessionExpiryInterval; sei == nil {
				c.prop.SessionExpiryInterval = 0
			} else if *sei < c.prop.SessionExpiryInterval {
				c.prop.SessionExpiryInterval = *sei
			}

			if mps := pc.Properties.MaximumPacketSize; mps == nil {
				c.prop.MaximumPacketSize = math.MaxUint32
			} else {
				c.prop.MaximumPacketSize = *mps
			}

			if tam := pc.Properties.TopicAliasMaximum; tam != nil {
				c.prop.TopicAliasMaximum = *tam
			}

			ack.Properties = &packets.Properties{
				RetainAvailable:       boolToByte(Cfg.Mqtt.RetainAvailable),
				SessionExpiryInterval: &c.prop.SessionExpiryInterval,
				ReceiveMaximum:        &Cfg.Mqtt.ReceiveMaximum,
				MaximumPacketSize:     &Cfg.Mqtt.MaximumPacketSize,
				ServerKeepAlive:       &c.prop.KeepAlive,
				TopicAliasMaximum:     &Cfg.Mqtt.TopicAliasMaximum,
				MaximumQoS:            &Cfg.Mqtt.MaximumQoS,
				WildcardSubAvailable:  boolToByte(Cfg.Mqtt.WildcardSub),
				SubIDAvailable:        boolToByte(Cfg.Mqtt.SubID),
				SharedSubAvailable:    boolToByte(Cfg.Mqtt.SharedSub),
			}
		} else {
			c.prop.KeepAlive = pc.KeepAlive
		}

		err := c.writePacket(ack)
		if err == nil {
			fmt.Println(*c.prop)
			log.Debugf("connect: connected %s", c.ID)
			return true
		}
		return false
	}
	return false
}

func (c *client) connectHandler(pc *packets.Connect) byte {
	return packets.Success
}

func boolToByte(b bool) *byte {
	var pb byte
	if b {
		pb = 1
	} else {
		pb = 0
	}
	return &pb
}

// Handle disconnect
func (c *client) disconnect(pd *packets.Disconnect) {
	c.close()
}

// Handle publish
func (c *client) publish(pp *packets.Publish) {
	fmt.Println(string(pp.Payload))
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
	fmt.Println("ping")
	resp := &packets.Pingresp{}
	_ = resp.Pack(c.conn)
}
