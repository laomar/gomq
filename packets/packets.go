package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Packet type
const (
	RESERVED = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH
)

// MQTT Version
const (
	V31 = iota + 3
	V311
	V5
)

// V3 reason code
const (
	Accepted                     = 0x00
	RefusedBadProtocolVersion    = 0x01
	RefusedIDRejected            = 0x02
	RefusedServerUnavailable     = 0x03
	RefusedBadUsernameOrPassword = 0x04
	RefusedNotAuthorised         = 0x05
	NetworkError                 = 0xFE
	ProtocolViolation            = 0xFF
)

// V5 reason code
const (
	Success                     = 0x00
	NormalDisconnection         = 0x00
	GrantedQoS0                 = 0x00
	GrantedQoS1                 = 0x01
	GrantedQoS2                 = 0x02
	DisconnectWithWillMessage   = 0x04
	NotMatchingSubscribers      = 0x10
	NoSubscriptionExisted       = 0x11
	ContinueAuthentication      = 0x18
	ReAuthenticate              = 0x19
	UnspecifiedError            = 0x80
	MalformedPacket             = 0x81
	ProtocolError               = 0x82
	ImplementationSpecificError = 0x83
	UnsupportedProtocolVersion  = 0x84
	ClientIdentifierNotValid    = 0x85
	BadUserNameOrPassword       = 0x86
	NotAuthorized               = 0x87
	ServerUnavailable           = 0x88
	ServerBusy                  = 0x89
	Banned                      = 0x8A
	BadAuthMethod               = 0x8C
	KeepAliveTimeout            = 0x8D
	SessionTakenOver            = 0x8E
	TopicFilterInvalid          = 0x8F
	TopicNameInvalid            = 0x90
	PacketIDInUse               = 0x91
	PacketIDNotFound            = 0x92
	RecvMaxExceeded             = 0x93
	TopicAliasInvalid           = 0x94
	PacketTooLarge              = 0x95
	MessageRateTooHigh          = 0x96
	QuotaExceeded               = 0x97
	AdminAction                 = 0x98
	PayloadFormatInvalid        = 0x99
	RetainNotSupported          = 0x9A
	QoSNotSupported             = 0x9B
	UseAnotherServer            = 0x9C
	ServerMoved                 = 0x9D
	SharedSubNotSupported       = 0x9E
	ConnectionRateExceeded      = 0x9F
	MaxConnectTime              = 0xA0
	SubIDNotSupported           = 0xA1
	WildcardSubNotSupported     = 0xA2
)

// FixHeader Struct
type FixHeader struct {
	PacketType byte
	Flags      byte
	Dup        bool
	Qos        byte
	Retain     bool
	RemainLen  int
}

func NewPacket(fh *FixHeader) Packet {
	switch fh.PacketType {
	case CONNECT:
		return &Connect{FixHeader: fh}
	case PINGREQ:
		return &Pingreq{FixHeader: fh}
	default:
		return nil
	}
}

func ReadPacket(r io.Reader) (Packet, error) {
	var fh FixHeader
	if err := fh.unpack(r); err != nil {
		return nil, err
	}
	p := NewPacket(&fh)
	if p == nil {
		return p, errors.New("nil")
	}
	err := p.Unpack(r)
	return p, err
}

type Packet interface {
	Pack(w io.Writer) error
	Unpack(r io.Reader) error
}

// Pack Fix Header
func (fh *FixHeader) pack(w io.Writer) error {
	if fh.PacketType == PUBLISH {
		if fh.Dup {
			fh.Flags |= 1 << 3
		}
		fh.Flags |= fh.Qos << 1
		if fh.Retain {
			fh.Flags |= 1
		}
	}
	b := make([]byte, 1)
	b[0] = fh.PacketType<<4 | fh.Flags
	b = append(b, encodeLength(fh.RemainLen)...)
	_, err := w.Write(b)
	return err
}

// Unpack Fix Header
func (fh *FixHeader) unpack(r io.Reader) error {
	var err error
	b := make([]byte, 1)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	fh.PacketType = b[0] >> 4
	fh.Flags = b[0] & 0x0F
	if fh.PacketType == PUBLISH {
		fh.Dup = fh.Flags>>3 > 0
		fh.Qos = fh.Flags >> 1 & 0x03
		fh.Retain = fh.Flags&0x01 > 0
	}
	if fh.RemainLen, err = decodeLength(r); err != nil {
		return err
	}
	return nil
}

// Decode Length to int
func decodeLength(r io.Reader) (int, error) {
	var l uint32
	var mul uint32
	buf := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, buf)
		if err != nil {
			return 0, err
		}
		l |= uint32(buf[0]&0x7F) << mul
		if buf[0]&0x80 == 0 {
			break
		}
		mul += 7
	}
	return int(l), nil
}

// Encode Length to []byte
func encodeLength(l int) []byte {
	buf := make([]byte, 0)
	for {
		b := byte(l % 128)
		l /= 128
		if l > 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if l <= 0 {
			break
		}
	}
	return buf
}

// Encode String to []byte
func encodeString(s string) []byte {
	l := len(s)
	buf := make([]byte, 2, 2+l)
	binary.BigEndian.PutUint16(buf, uint16(l))
	buf = append(buf, s...)
	return buf
}

// Decode []byte to String
func decodeString(r *bytes.Buffer) string {
	l := int(binary.BigEndian.Uint16(r.Next(2)))
	s := string(r.Next(l))
	return s
}

// Read Buffer to uint
func readUint16(r *bytes.Buffer) uint16 {
	return binary.BigEndian.Uint16(r.Next(2))
}
func readUint32(r *bytes.Buffer) uint32 {
	return binary.BigEndian.Uint32(r.Next(4))
}

// Write uint to Buffer
func writeUint16(w *bytes.Buffer, i uint16) {
	w.WriteByte(byte(i >> 8))
	w.WriteByte(byte(i))
}
func writeUint32(w *bytes.Buffer, i uint32) {
	w.WriteByte(byte(i >> 24))
	w.WriteByte(byte(i >> 16))
	w.WriteByte(byte(i >> 8))
	w.WriteByte(byte(i))
}
