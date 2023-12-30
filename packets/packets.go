package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Packet type
const (
	CONNECT = iota + 1
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

// MQTT Qos
const (
	Qos0 = iota
	Qos1
	Qos2
)

// V3 reason code
const (
	Accepted                     = 0x00
	RefusedBadProtocolVersion    = 0x01
	RefusedIDRejected            = 0x02
	RefusedServerUnavailable     = 0x03
	RefusedBadUsernameOrPassword = 0x04
	RefusedNotAuthorised         = 0x05
	SubFail                      = 0x80
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

var (
	ErrMalformed = &Error{Code: MalformedPacket}
	ErrProtocol  = &Error{Code: ProtocolError}
)

type Error struct {
	Code   byte
	Reason string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error Code: %x, Reason: %s", e.Code, e.Reason)
}

func NewPacket(fh *FixHeader, v byte) Packet {
	switch fh.PacketType {
	case CONNECT:
		return &Connect{FixHeader: fh}
	case PINGREQ:
		return &Pingreq{FixHeader: fh}
	case PUBLISH:
		return &Publish{FixHeader: fh, Version: v}
	case PUBREL:
		return &Pubrel{FixHeader: fh, Version: v}
	case DISCONNECT:
		return &Disconnect{FixHeader: fh, Version: v}
	case SUBSCRIBE:
		return &Subscribe{FixHeader: fh, Version: v}
	case UNSUBSCRIBE:
		return &Unsubscribe{FixHeader: fh, Version: v}
	case AUTH:
		return &Auth{FixHeader: fh}
	default:
		return nil
	}
}

// Packet interface
type Packet interface {
	Pack(w io.Writer) error
	Unpack(r io.Reader) error
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
	if r.Len() < 2 {
		return ""
	}
	l := int(binary.BigEndian.Uint16(r.Next(2)))
	s := string(r.Next(l))
	return s
}

// Read Buffer to uint
func readUint16(r *bytes.Buffer) uint16 {
	if r.Len() < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(r.Next(2))
}
func readUint32(r *bytes.Buffer) uint32 {
	if r.Len() < 4 {
		return 0
	}
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
