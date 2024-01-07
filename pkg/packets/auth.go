package packets

import (
	"bytes"
	"io"
)

type Auth struct {
	FixHeader  *FixHeader
	ReasonCode byte
	Properties *Properties
}

// Pack Auth Packet
func (a *Auth) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	bufw.WriteByte(a.ReasonCode)
	if a.Properties != nil {
		a.Properties.Pack(bufw)
	}
	a.FixHeader = &FixHeader{
		PacketType: AUTH,
		RemainLen:  bufw.Len(),
	}
	if err := a.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Auth Packet
func (a *Auth) Unpack(r io.Reader) error {
	var err error
	buf := make([]byte, a.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	if a.ReasonCode, err = bufr.ReadByte(); err != nil {
		return ErrMalformed
	}
	a.Properties = &Properties{}
	if err = a.Properties.Unpack(bufr); err != nil {
		return err
	}
	return nil
}
