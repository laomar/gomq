package packets

import (
	"bytes"
	"io"
)

type Publish struct {
	FixHeader  *FixHeader
	Version    byte
	TopicName  string
	PacketID   uint16
	Properties *Properties
	Payload    []byte
}

// Pack Publish Packet
func (p *Publish) Pack(w io.Writer) error {
	return nil
}

// Unpack Publish Packet
func (p *Publish) Unpack(r io.Reader) error {
	var err error
	buf := make([]byte, p.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	p.TopicName = decodeString(bufr)
	if p.FixHeader.Qos > Qos0 {
		p.PacketID = readUint16(bufr)
	}
	if p.Version == V5 {
		p.Properties = &Properties{}
		if err = p.Properties.Unpack(bufr); err != nil {
			return err
		}
	}
	p.Payload = bufr.Next(bufr.Len())
	return nil
}
