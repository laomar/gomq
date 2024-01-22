package server

import (
	"context"
	. "github.com/laomar/gomq/config"
	"github.com/laomar/gomq/log"
	"github.com/laomar/gomq/pkg/packets"
	"math"
	"net"
	"strings"
	"time"
)

const (
	Connecting = iota
	Connected
	Disconnected
)

// Client
type ClientProp struct {
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
type Client struct {
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
	prop    *ClientProp
}

func (c *Client) serve() {
	go c.readLoop()
	if c.connect() {
		go c.writeLoop()
		go c.handleLoop()
	}
	<-c.ctx.Done()
}

// Read packet
func (c *Client) readPacket() (packets.Packet, error) {
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
func (c *Client) writePacket(p packets.Packet) error {
	return p.Pack(c.conn)
}

func (c *Client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("client read recover: %v %v", c.ID, r)
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
				log.Debugf("client: %v", err)
				c.close()
				return
			}
			c.in <- p
		}
	}
}

func (c *Client) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("client write recover: %v %v", c.ID, r)
		}
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case out := <-c.out:
			err := c.writePacket(out)
			if err != nil {
				log.Debugf("client: %v cid=%v", err, c.ID)
				c.close()
				return
			}
		}
	}
}

func (c *Client) handleLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("client handle recover: %v %v", c.ID, r)
		}
	}()
	for in := range c.in {
		switch p := in.(type) {
		case *packets.Disconnect:
			c.disconnect(p)
		case *packets.Pingreq:
			c.pingReqHandler()
		case *packets.Publish:
			c.publishHandler(p)
		case *packets.Pubrel:
			c.pubrel(p)
		case *packets.Subscribe:
			c.subscribeHandler(p)
		case *packets.Unsubscribe:
			c.unsubscribeHandler(p)
		case *packets.Auth:
			c.auth(p)
		}
	}
}

// Client close
func (c *Client) close() {
	defer c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
	//c.server.topicStore.UnsubscribeAll(c.ID)
}

// Handle connect
func (c *Client) connect() bool {
	var pc *packets.Connect
	for in := range c.in {
		var code byte
		switch p := in.(type) {
		case *packets.Connect:
			pc = p
			c.Version = pc.Version
			if len(pc.ClientID) == 0 {
				code = packets.ClientIdentifierNotValid
				break
			}
			code = c.connectHandler(pc)
		case *packets.Auth:
		default:
			code = packets.MalformedPacket
		}

		// continue authentication
		if code == packets.ContinueAuthentication {
			auth := &packets.Auth{
				ReasonCode: code,
				Properties: &packets.Properties{
					AuthMethod: pc.Properties.AuthMethod,
				},
			}
			_ = c.writePacket(auth)
			continue
		}

		ack := &packets.Connack{
			Version:    c.Version,
			ReasonCode: code,
		}

		// connect fail
		if code != packets.Success {
			_ = c.writePacket(ack)
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
				TopicAliasMaximum:     &Cfg.Mqtt.MaxTopicAlias,
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
			log.Debugf("mqtt: connected cid=%s addr=%s", c.ID, c.conn.RemoteAddr())
			return true
		}
		return false
	}
	return false
}

func (c *Client) connectHandler(pc *packets.Connect) byte {
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
func (c *Client) disconnect(pd *packets.Disconnect) {
	c.close()
}

// Handle publish
func (c *Client) publishHandler(pp *packets.Publish) {
	if !Cfg.Mqtt.RetainAvailable && pp.FixHeader.Retain {
	}

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
func (c *Client) pubrel(pp *packets.Pubrel) {
	rec := &packets.Pubcomp{
		Version:    c.Version,
		ReasonCode: packets.Success,
		PacketID:   pp.PacketID,
	}
	_ = rec.Pack(c.conn)
}

// Handle Subscribe
func (c *Client) subscribeHandler(ps *packets.Subscribe) {
	var err error
	suback := &packets.Suback{
		Version:  c.Version,
		PacketID: ps.PacketID,
		Payload:  make([]byte, len(ps.Subscriptions)),
	}

	var subid uint32
	if c.Version == packets.V5 {
		if Cfg.Mqtt.SubID && len(ps.Properties.SubscriptionIdentifier) > 0 {
			subid = ps.Properties.SubscriptionIdentifier[0]
		}
	}

	isExist := false
	for i, subscription := range ps.Subscriptions {
		if subscription.Qos > Cfg.Mqtt.MaximumQoS {
			subscription.Qos = Cfg.Mqtt.MaximumQoS
		}

		topics := strings.Split(subscription.Topic, "/")
		if len(topics) > int(Cfg.Mqtt.MaxTopicLevel) && Cfg.Mqtt.MaxTopicLevel > 0 {
			suback.Payload[i] = packets.Code(c.Version, packets.TopicFilterInvalid)
			continue
		}
		if len(topics) >= 2 && topics[0] == "$share" {
			subscription.ShareName = topics[1]
		}

		if !Cfg.Mqtt.SharedSub && subscription.ShareName != "" {
			suback.Payload[i] = packets.Code(c.Version, packets.SharedSubNotSupported)
			continue
		}
		if !Cfg.Mqtt.WildcardSub {
			if strings.Contains(subscription.Topic, "#") || strings.Contains(subscription.Topic, "+") {
				suback.Payload[i] = packets.Code(c.Version, packets.WildcardSubNotSupported)
				continue
			}
		}
		if !Cfg.Mqtt.SubID && subid > 0 {
			suback.Payload[i] = packets.Code(c.Version, packets.SubIDNotSupported)
			continue
		}

		subscription.SubID = subid
		isExist, err = c.server.topicStore.Subscribe(c.ID, &subscription)
		if err != nil {
			log.Debugf("subscribe: failed cid=%s topic=%s %s", c.ID, subscription.Topic, err)
			suback.Payload[i] = packets.UnspecifiedError
			continue
		}
		suback.Payload[i] = subscription.Qos
		log.Debugf("subscribe: succeed cid=%s topic=%s", c.ID, subscription.Topic)

		if subscription.ShareName != "" || isExist || subscription.RetainHandling >= 2 {
			continue
		}

		c.server.cluster.Subscribe(c.ID, subscription.Topic)

		// 保留信息
	}
	_ = c.writePacket(suback)
}

// Handle Unsubscribe
func (c *Client) unsubscribeHandler(pu *packets.Unsubscribe) {
	for _, topic := range pu.Topics {
		c.server.topicStore.Unsubscribe(c.ID, topic)
	}
	ack := &packets.Unsuback{
		Version:  c.Version,
		PacketID: pu.PacketID,
		Payload:  []byte{packets.Success},
	}
	_ = c.writePacket(ack)
}

// Handle auth
func (c *Client) auth(pa *packets.Auth) {

}

// Handle ping
func (c *Client) pingReqHandler() {
	resp := &packets.Pingresp{}
	_ = c.writePacket(resp)
}
