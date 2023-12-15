package packets

import (
	"bytes"
	"io"
)

type Topic struct {
	Name              string
	RetainHandling    byte
	RetainAsPublished bool
	NoLocal           bool
	Qos               byte
}
type Subscribe struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	Topics     []Topic
	Properties *Properties
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
		topic := Topic{Name: name}
		opts, err := bufr.ReadByte()
		if err != nil {
			return ErrMalformed
		}
		if s.Version == V5 {
			topic.RetainHandling = (opts >> 4) & 0x03
			topic.RetainAsPublished = opts&0x08 > 0
			topic.NoLocal = opts&0x04 > 0
			topic.Qos = opts & 0x03
		} else {
			topic.Qos = opts
		}
		if topic.Qos > Qos2 {
			return ErrProtocol
		}
		s.Topics = append(s.Topics, topic)
	}
}
