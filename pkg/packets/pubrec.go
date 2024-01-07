package packets

import (
	"bytes"
	"io"
)

// Pubrec Packet
type Pubrec struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	ReasonCode byte
	Properties *Properties
}

// Pack Pubrec Packet
func (p *Pubrec) Pack(w io.Writer) error {
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
		PacketType: PUBREC,
		RemainLen:  bufw.Len(),
	}
	if err := p.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Pubrec Packet
func (p *Pubrec) Unpack(r io.Reader) error {
	return nil
}
