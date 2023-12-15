package packets

import (
	"bytes"
	"io"
)

// Unsuback Packet
type Unsuback struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	Properties *Properties
	Payload    []byte
}

// Pack Unsuback Packet
func (u *Unsuback) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	writeUint16(bufw, u.PacketID)
	if u.Version == V5 {
		if u.Properties != nil {
			u.Properties.Pack(bufw)
		} else {
			bufw.WriteByte(0)
		}
	}
	bufw.Write(u.Payload)
	u.FixHeader = &FixHeader{
		PacketType: UNSUBACK,
		RemainLen:  bufw.Len(),
	}
	if err := u.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Unsuback Packet
func (u *Unsuback) Unpack(r io.Reader) error {
	return nil
}
