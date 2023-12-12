package packets

import (
	"bytes"
	"encoding/binary"
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
	err := p.unpack(r)
	return p, err
}

type Packet interface {
	pack(w io.Writer) error
	unpack(r io.Reader) error
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
	b := make([]byte, 2)
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
	buf := make([]byte, 1)
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
