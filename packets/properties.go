package packets

import (
	"bytes"
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
	PayloadFormat          *byte
	MessageExpiry          *uint32
	ContentType            string
	ResponseTopic          string
	CorrelationData        string
	SubscriptionIdentifier []uint32
	SessionExpiryInterval  *uint32
	AssignedClientID       string
	ServerKeepAlive        *uint16
	AuthMethod             string
	AuthData               string
	RequestProblemInfo     *byte
	WillDelayInterval      *uint32
	RequestResponseInfo    *byte
	ResponseInfo           string
	ServerReference        string
	ReasonString           string
	ReceiveMaximum         *uint16
	TopicAliasMaximum      *uint16
	TopicAlias             *uint16
	MaximumQoS             *byte
	RetainAvailable        *byte
	User                   map[string]string
	MaximumPacketSize      *uint32
	WildcardSubAvailable   *byte
	SubIDAvailable         *byte
	SharedSubAvailable     *byte
}

// Pack Connect Packet
func (p *Properties) pack(w *bytes.Buffer) {
	if p.PayloadFormat != nil {
		w.WriteByte(PayloadFormat)
		w.WriteByte(*p.PayloadFormat)
	}
	if p.MessageExpiry != nil {
		w.WriteByte(MessageExpiry)
		writeUint32(w, *p.MessageExpiry)
	}
	if p.ContentType != "" {
		w.WriteByte(ContentType)
		w.Write([]byte(p.ContentType))
	}
	if p.ResponseTopic != "" {
		w.WriteByte(ResponseTopic)
		w.Write([]byte(p.ResponseTopic))
	}
	if p.CorrelationData != "" {
		w.WriteByte(CorrelationData)
		w.Write([]byte(p.CorrelationData))
	}
	if len(p.SubscriptionIdentifier) > 0 {
		for _, si := range p.SubscriptionIdentifier {
			w.WriteByte(SubscriptionIdentifier)
			w.Write(encodeLength(int(si)))
		}
	}
	if p.SessionExpiryInterval != nil {
		w.WriteByte(SessionExpiryInterval)
		writeUint32(w, *p.SessionExpiryInterval)
	}
	if p.AssignedClientID != "" {
		w.WriteByte(AssignedClientID)
		w.Write([]byte(p.AssignedClientID))
	}
	if p.ServerKeepAlive != nil {
		w.WriteByte(ServerKeepAlive)
		writeUint16(w, *p.ServerKeepAlive)
	}
	if p.AuthMethod != "" {
		w.WriteByte(AuthMethod)
		w.Write([]byte(p.AuthMethod))
	}
	if p.AuthData != "" {
		w.WriteByte(AuthData)
		w.Write([]byte(p.AuthData))
	}
	if p.RequestProblemInfo != nil {
		w.WriteByte(RequestProblemInfo)
		w.WriteByte(*p.RequestProblemInfo)
	}
	if p.WillDelayInterval != nil {
		w.WriteByte(WillDelayInterval)
		writeUint32(w, *p.WillDelayInterval)
	}
	if p.RequestResponseInfo != nil {
		w.WriteByte(RequestResponseInfo)
		w.WriteByte(*p.RequestResponseInfo)
	}
	if p.ResponseInfo != "" {
		w.WriteByte(ResponseInfo)
		w.Write([]byte(p.ResponseInfo))
	}
	if p.ServerReference != "" {
		w.WriteByte(ServerReference)
		w.Write([]byte(p.ServerReference))
	}
	if p.ReasonString != "" {
		w.WriteByte(ReasonString)
		w.Write([]byte(p.ReasonString))
	}
	if p.ReceiveMaximum != nil {
		w.WriteByte(ReceiveMaximum)
		writeUint16(w, *p.ReceiveMaximum)
	}
	if p.TopicAliasMaximum != nil {
		w.WriteByte(TopicAliasMaximum)
		writeUint16(w, *p.TopicAliasMaximum)
	}
	if p.TopicAlias != nil {
		w.WriteByte(TopicAlias)
		writeUint16(w, *p.TopicAlias)
	}
	if p.MaximumQoS != nil {
		w.WriteByte(MaximumQoS)
		w.WriteByte(*p.MaximumQoS)
	}
	if p.RetainAvailable != nil {
		w.WriteByte(RetainAvailable)
		w.WriteByte(*p.RetainAvailable)
	}
	for k, v := range p.User {
		w.WriteByte(User)
		w.Write(encodeString(k))
		w.Write(encodeString(v))
	}
	if p.MaximumPacketSize != nil {
		w.WriteByte(MaximumPacketSize)
		writeUint32(w, *p.MaximumPacketSize)
	}
	if p.WildcardSubAvailable != nil {
		w.WriteByte(WildcardSubAvailable)
		w.WriteByte(*p.WildcardSubAvailable)
	}
	if p.SubIDAvailable != nil {
		w.WriteByte(SubIDAvailable)
		w.WriteByte(*p.SubIDAvailable)
	}
	if p.SharedSubAvailable != nil {
		w.WriteByte(SharedSubAvailable)
		w.WriteByte(*p.SharedSubAvailable)
	}
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
			u := readUint32(buf)
			p.MessageExpiry = &u
		case ContentType:
			p.ContentType = decodeString(buf)
		case ResponseTopic:
			p.ResponseTopic = decodeString(buf)
		case CorrelationData:
			p.CorrelationData = decodeString(buf)
		case SubscriptionIdentifier:
			si, err := decodeLength(buf)
			if err != nil {
				return err
			}
			p.SubscriptionIdentifier = append(p.SubscriptionIdentifier, uint32(si))
		case SessionExpiryInterval:
			u := readUint32(buf)
			p.SessionExpiryInterval = &u
		case AssignedClientID:
			p.AssignedClientID = decodeString(buf)
		case ServerKeepAlive:
			u := readUint16(buf)
			p.ServerKeepAlive = &u
		case AuthMethod:
			p.AuthMethod = decodeString(buf)
		case AuthData:
			p.AuthData = decodeString(buf)
		case RequestProblemInfo:
			b, err = buf.ReadByte()
			p.RequestProblemInfo = &b
		case WillDelayInterval:
			u := readUint32(buf)
			p.WillDelayInterval = &u
		case RequestResponseInfo:
			b, err = buf.ReadByte()
			p.RequestResponseInfo = &b
		case ResponseInfo:
			p.ResponseInfo = decodeString(buf)
		case ServerReference:
			p.ServerReference = decodeString(buf)
		case ReasonString:
			p.ReasonString = decodeString(buf)
		case ReceiveMaximum:
			u := readUint16(buf)
			p.ReceiveMaximum = &u
		case TopicAliasMaximum:
			u := readUint16(buf)
			p.TopicAliasMaximum = &u
		case TopicAlias:
			u := readUint16(buf)
			p.TopicAlias = &u
		case MaximumQoS:
			b, err = buf.ReadByte()
			p.MaximumQoS = &b
		case RetainAvailable:
			b, err = buf.ReadByte()
			p.RetainAvailable = &b
		case User:
			if p.User == nil {
				p.User = make(map[string]string)
			}
			k := decodeString(buf)
			v := decodeString(buf)
			p.User[k] = v
		case MaximumPacketSize:
			u := readUint32(buf)
			p.MaximumPacketSize = &u
		case WildcardSubAvailable:
			b, err = buf.ReadByte()
			p.WildcardSubAvailable = &b
		case SubIDAvailable:
			b, err = buf.ReadByte()
			p.SubIDAvailable = &b
		case SharedSubAvailable:
			b, err = buf.ReadByte()
			p.SharedSubAvailable = &b
		default:
			fmt.Println(prop)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
