package packets

import (
	"bytes"
	"io"
)

// Connack Packet
type Connack struct {
	FixHeader      *FixHeader
	Version        byte
	SessionPresent bool
	ReasonCode     byte
	Properties     *Properties
}

// Pack Connack Packet
func (c *Connack) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	if c.SessionPresent {
		bufw.WriteByte(1)
	} else {
		bufw.WriteByte(0)
	}
	bufw.WriteByte(c.ReasonCode)
	if c.Version == V5 && c.Properties != nil {
		if err := c.Properties.Pack(bufw); err != nil {
			return err
		}
	}
	c.FixHeader = &FixHeader{
		PacketType: CONNACK,
		RemainLen:  bufw.Len(),
	}
	if err := c.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Connack Packet
func (c *Connack) Unpack(r io.Reader) error {
	return nil
}
