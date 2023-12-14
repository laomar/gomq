package packets

import (
	"bytes"
	"io"
)

// Disconnect Packet
type Disconnect struct {
	FixHeader  *FixHeader
	Version    byte
	ReasonCode byte
	Properties *Properties
}

// Pack Disconnect Packet
func (c *Disconnect) Pack(w io.Writer) error {
	return nil
}

// Unpack Disconnect Packet
func (c *Disconnect) Unpack(r io.Reader) error {
	var err error
	buf := make([]byte, c.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	if c.Version == V5 {
		bufr := bytes.NewBuffer(buf)
		if c.ReasonCode, err = bufr.ReadByte(); err != nil {
			return err
		}
		c.Properties = &Properties{}
		return c.Properties.Unpack(bufr)
	}
	return nil
}
