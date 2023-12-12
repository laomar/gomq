package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Properties
const (
	PayloadFormat          = 0x01
	MessageExpiry          = 0x02
	ContentType            = 0x03
	ResponseTopic          = 0x08
	CorrelationData        = 0x09
	SubscriptionIdentifier = 0x0B
	SessionExpiryInterval  = 0x11
	AssignedClientID       = 0x12
	ServerKeepAlive        = 0x13
	AuthMethod             = 0x15
	AuthData               = 0x16
	RequestProblemInfo     = 0x17
	WillDelayInterval      = 0x18
	RequestResponseInfo    = 0x19
	ResponseInfo           = 0x1A
	ServerReference        = 0x1C
	ReasonString           = 0x1F
	ReceiveMaximum         = 0x21
	TopicAliasMaximum      = 0x22
	TopicAlias             = 0x23
	MaximumQoS             = 0x24
	RetainAvailable        = 0x25
	User                   = 0x26
	MaximumPacketSize      = 0x27
	WildcardSubAvailable   = 0x28
	SubIDAvailable         = 0x29
	SharedSubAvailable     = 0x2A
)

type Properties struct {
	PayloadFormat         *byte
	MessageExpiry         *uint32
	ContentType           string
	ResponseTopic         string
	CorrelationData       string
	SessionExpiryInterval *uint32
	RequestProblemInfo    *byte
	WillDelayInterval     *uint32
	RequestResponseInfo   *byte
	ReceiveMaximum        *uint16
	TopicAliasMaximum     *uint16
	User                  []map[string]string
	MaximumPacketSize     *uint32
}

// Unpack Properties
func (p *Properties) unpack(r *bytes.Buffer) error {
	var err error
	l, err := decodeLength(r)
	if err != nil {
		return err
	}
	if l == 0 {
		return nil
	}
	buf := bytes.NewBuffer(r.Next(l))
	var prop, b byte
	for {
		prop, err = buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		switch prop {
		case PayloadFormat:
			b, err = buf.ReadByte()
			p.PayloadFormat = &b
		case MessageExpiry:
			u := binary.BigEndian.Uint32(buf.Next(4))
			p.MessageExpiry = &u
		case ContentType:
			p.ContentType = decodeString(buf)
		case ResponseTopic:
			p.ResponseTopic = decodeString(buf)
		case CorrelationData:
			p.CorrelationData = decodeString(buf)
		case SessionExpiryInterval:
			u := binary.BigEndian.Uint32(buf.Next(4))
			p.SessionExpiryInterval = &u
		case RequestProblemInfo:
			b, err = buf.ReadByte()
			p.RequestProblemInfo = &b
		case WillDelayInterval:
			u := binary.BigEndian.Uint32(buf.Next(4))
			p.WillDelayInterval = &u
		case RequestResponseInfo:
			b, err = buf.ReadByte()
			p.RequestResponseInfo = &b
		case ReceiveMaximum:
			u := binary.BigEndian.Uint16(buf.Next(2))
			p.ReceiveMaximum = &u
		case TopicAliasMaximum:
			u := binary.BigEndian.Uint16(buf.Next(2))
			p.TopicAliasMaximum = &u
		case User:
			k := decodeString(buf)
			v := decodeString(buf)
			p.User = append(p.User, map[string]string{k: v})
		case MaximumPacketSize:
			u := binary.BigEndian.Uint32(buf.Next(4))
			p.MaximumPacketSize = &u
		default:
			fmt.Println(prop)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
