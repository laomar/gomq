package packets

import "io"

// Pingresp packet
type Pingresp struct {
	FixHeader *FixHeader
}

// Pack Pingresp Packet
func (c *Pingresp) Pack(w io.Writer) error {
	c.FixHeader = &FixHeader{
		PacketType: PINGRESP,
		RemainLen:  0,
	}
	return c.FixHeader.pack(w)
}

// Unpack Pingresp Packet
func (c *Pingresp) Unpack(r io.Reader) error {
	return nil
}
