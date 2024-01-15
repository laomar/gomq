package packets

import (
	"bytes"
	"io"
)

type Subscription struct {
	Topic             string
	ShareName         string
	RetainHandling    byte
	RetainAsPublished bool
	NoLocal           bool
	Qos               byte
	SubID             uint32
}
type Subscribe struct {
	FixHeader     *FixHeader
	Version       byte
	PacketID      uint16
	Subscriptions []Subscription
	Properties    *Properties
}

// Pack Subscribe Packet
func (s *Subscribe) Pack(w io.Writer) error {
	return nil
}

// Unpack Subscribe Packet
func (s *Subscribe) Unpack(r io.Reader) error {
	buf := make([]byte, s.FixHeader.RemainLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	s.PacketID = readUint16(bufr)
	if s.Version == V5 {
		s.Properties = &Properties{}
		if err := s.Properties.Unpack(bufr); err != nil {
			return err
		}
	}
	for {
		if bufr.Len() == 0 {
			return nil
		}
		name := decodeString(bufr)
		sub := Subscription{Topic: name}
		opts, err := bufr.ReadByte()
		if err != nil {
			return ErrMalformed
		}
		if s.Version == V5 {
			sub.RetainHandling = (opts >> 4) & 0x03
			sub.RetainAsPublished = opts&0x08 > 0
			sub.NoLocal = opts&0x04 > 0
			sub.Qos = opts & 0x03
		} else {
			sub.Qos = opts
		}
		if sub.Qos > Qos2 {
			return ErrProtocol
		}
		s.Subscriptions = append(s.Subscriptions, sub)
	}
}
