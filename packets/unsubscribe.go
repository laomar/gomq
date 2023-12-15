package packets

import (
	"bytes"
	"io"
)

type Unsubscribe struct {
	FixHeader  *FixHeader
	Version    byte
	PacketID   uint16
	Topics     []string
	Properties *Properties
}

// Pack Unsubscribe Packet
func (u *Unsubscribe) Pack(w io.Writer) error {
	return nil
}

// Unpack Unsubscribe Packet
func (u *Unsubscribe) Unpack(r io.Reader) error {
	buf := make([]byte, u.FixHeader.RemainLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	u.PacketID = readUint16(bufr)
	if u.Version == V5 {
		u.Properties = &Properties{}
		if err := u.Properties.Unpack(bufr); err != nil {
			return err
		}
	}
	for {
		if bufr.Len() == 0 {
			return nil
		}
		topic := decodeString(bufr)
		u.Topics = append(u.Topics, topic)
	}
}
