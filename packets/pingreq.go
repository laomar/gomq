package packets

import "io"

// Pingreq packet
type Pingreq struct {
	FixHeader *FixHeader
}

// Pack Pingreq Packet
func (c *Pingreq) Pack(w io.Writer) error {
	c.FixHeader = &FixHeader{
		PacketType: PINGREQ,
		RemainLen:  0,
	}
	return c.FixHeader.pack(w)
}

// Unpack Pingreq Packet
func (c *Pingreq) Unpack(r io.Reader) error {
	return nil
}
