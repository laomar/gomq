package packets

import (
	"bytes"
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
func (p *Properties) Pack(w *bytes.Buffer) error {
	buf := &bytes.Buffer{}
	if p.PayloadFormat != nil {
		buf.WriteByte(PayloadFormat)
		buf.WriteByte(*p.PayloadFormat)
	}
	if p.MessageExpiry != nil {
		buf.WriteByte(MessageExpiry)
		writeUint32(buf, *p.MessageExpiry)
	}
	if p.ContentType != "" {
		buf.WriteByte(ContentType)
		buf.Write([]byte(p.ContentType))
	}
	if p.ResponseTopic != "" {
		buf.WriteByte(ResponseTopic)
		buf.Write([]byte(p.ResponseTopic))
	}
	if p.CorrelationData != "" {
		buf.WriteByte(CorrelationData)
		buf.Write([]byte(p.CorrelationData))
	}
	if len(p.SubscriptionIdentifier) > 0 {
		for _, si := range p.SubscriptionIdentifier {
			buf.WriteByte(SubscriptionIdentifier)
			buf.Write(encodeLength(int(si)))
		}
	}
	if p.SessionExpiryInterval != nil {
		buf.WriteByte(SessionExpiryInterval)
		writeUint32(buf, *p.SessionExpiryInterval)
	}
	if p.AssignedClientID != "" {
		buf.WriteByte(AssignedClientID)
		buf.Write([]byte(p.AssignedClientID))
	}
	if p.ServerKeepAlive != nil {
		buf.WriteByte(ServerKeepAlive)
		writeUint16(buf, *p.ServerKeepAlive)
	}
	if p.AuthMethod != "" {
		buf.WriteByte(AuthMethod)
		buf.Write([]byte(p.AuthMethod))
	}
	if p.AuthData != "" {
		buf.WriteByte(AuthData)
		buf.Write([]byte(p.AuthData))
	}
	if p.RequestProblemInfo != nil {
		buf.WriteByte(RequestProblemInfo)
		buf.WriteByte(*p.RequestProblemInfo)
	}
	if p.WillDelayInterval != nil {
		buf.WriteByte(WillDelayInterval)
		writeUint32(buf, *p.WillDelayInterval)
	}
	if p.RequestResponseInfo != nil {
		buf.WriteByte(RequestResponseInfo)
		buf.WriteByte(*p.RequestResponseInfo)
	}
	if p.ResponseInfo != "" {
		buf.WriteByte(ResponseInfo)
		buf.Write([]byte(p.ResponseInfo))
	}
	if p.ServerReference != "" {
		buf.WriteByte(ServerReference)
		buf.Write([]byte(p.ServerReference))
	}
	if p.ReasonString != "" {
		buf.WriteByte(ReasonString)
		buf.Write([]byte(p.ReasonString))
	}
	if p.ReceiveMaximum != nil {
		buf.WriteByte(ReceiveMaximum)
		writeUint16(buf, *p.ReceiveMaximum)
	}
	if p.TopicAliasMaximum != nil {
		buf.WriteByte(TopicAliasMaximum)
		writeUint16(buf, *p.TopicAliasMaximum)
	}
	if p.TopicAlias != nil {
		buf.WriteByte(TopicAlias)
		writeUint16(buf, *p.TopicAlias)
	}
	if p.MaximumQoS != nil {
		buf.WriteByte(MaximumQoS)
		buf.WriteByte(*p.MaximumQoS)
	}
	if p.RetainAvailable != nil {
		buf.WriteByte(RetainAvailable)
		buf.WriteByte(*p.RetainAvailable)
	}
	for k, v := range p.User {
		buf.WriteByte(User)
		buf.Write(encodeString(k))
		buf.Write(encodeString(v))
	}
	if p.MaximumPacketSize != nil {
		buf.WriteByte(MaximumPacketSize)
		writeUint32(buf, *p.MaximumPacketSize)
	}
	if p.WildcardSubAvailable != nil {
		buf.WriteByte(WildcardSubAvailable)
		buf.WriteByte(*p.WildcardSubAvailable)
	}
	if p.SubIDAvailable != nil {
		buf.WriteByte(SubIDAvailable)
		buf.WriteByte(*p.SubIDAvailable)
	}
	if p.SharedSubAvailable != nil {
		buf.WriteByte(SharedSubAvailable)
		buf.WriteByte(*p.SharedSubAvailable)
	}
	l := encodeLength(buf.Len())
	w.Write(l)
	_, err := buf.WriteTo(w)
	return err
}

// Unpack Properties
func (p *Properties) Unpack(r *bytes.Buffer) error {
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
		}
		if err != nil {
			return err
		}
	}
	return nil
}
