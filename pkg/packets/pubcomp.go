package packets

import (
	"bytes"
	"io"
)

// Pubcomp Packet
type Pubcomp struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	ReasonCode byte
	Properties *Properties
}

// Pack Pubcomp Packet
func (p *Pubcomp) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	writeUint16(bufw, p.PacketID)
	if p.Version == V5 {
		bufw.WriteByte(p.ReasonCode)
		if p.Properties != nil {
			p.Properties.Pack(bufw)
		} else {
			bufw.WriteByte(0)
		}
	}
	p.FixHeader = &FixHeader{
		PacketType: PUBCOMP,
		RemainLen:  bufw.Len(),
	}
	if err := p.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Pubcomp Packet
func (p *Pubcomp) Unpack(r io.Reader) error {
	return nil
}
