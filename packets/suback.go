package packets

import (
	"bytes"
	"io"
)

// Suback Packet
type Suback struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	Properties *Properties
	Payload    []byte
}

// Pack Suback Packet
func (s *Suback) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	writeUint16(bufw, s.PacketID)
	if s.Version == V5 {
		if s.Properties != nil {
			s.Properties.Pack(bufw)
		} else {
			bufw.WriteByte(0)
		}
	}
	bufw.Write(s.Payload)
	s.FixHeader = &FixHeader{
		PacketType: SUBACK,
		RemainLen:  bufw.Len(),
	}
	if err := s.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Suback Packet
func (s *Suback) Unpack(r io.Reader) error {
	return nil
}
