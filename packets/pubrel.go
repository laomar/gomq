package packets

import (
	"bytes"
	"io"
)

// Pubrel Packet
type Pubrel struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	ReasonCode byte
	Properties *Properties
}

// Pack Pubrel Packet
func (p *Pubrel) Pack(w io.Writer) error {
	return nil
}

// Unpack Pubrel Packet
func (p *Pubrel) Unpack(r io.Reader) error {
	var err error
	buf := make([]byte, p.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	p.PacketID = readUint16(bufr)
	if p.Version == V5 {
		p.Properties = &Properties{}
		if err = p.Properties.Unpack(bufr); err != nil {
			return err
		}
	}
	return nil
}
