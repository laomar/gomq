package packets

import (
	"bytes"
	"io"
)

// Puback Packet
type Puback struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	ReasonCode byte
	Properties *Properties
}

// Pack Puback Packet
func (p *Puback) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	writeUint16(bufw, p.PacketID)
	if p.Version == V5 && p.Properties != nil {
		bufw.WriteByte(p.ReasonCode)
		p.Properties.Pack(bufw)
	}
	p.FixHeader = &FixHeader{
		PacketType: PUBACK,
		RemainLen:  bufw.Len(),
	}
	if err := p.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Puback Packet
func (p *Puback) Unpack(r io.Reader) error {
	return nil
}
