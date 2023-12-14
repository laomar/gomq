package packets

import (
	"bytes"
	"io"
)

// Connect Packet
type Connect struct {
	FixHeader *FixHeader
	// Variable Header
	Protocol string
	Version  byte
	// Connect Flags
	UsernameFlag bool
	PasswordFlag bool
	WillRetain   bool
	WillQos      byte
	WillFlag     bool
	CleanStart   bool
	Reserved     bool
	KeepAlive    uint16
	Properties   *Properties
	// Payload
	ClientID       string
	WillProperties *Properties
	WillTopic      string
	WillMsg        string
	Username       string
	Password       string
}

// Pack Connect Packet
func (c *Connect) Pack(w io.Writer) error {
	bufw := &bytes.Buffer{}
	bufw.WriteByte(0x00)
	bufw.WriteByte(0x04)
	bufw.Write([]byte(c.Protocol))
	bufw.WriteByte(c.Version)
	// write flags
	var flags byte
	if c.UsernameFlag {
		flags |= 1 << 7
	}
	if c.PasswordFlag {
		flags |= 1 << 6
	}
	if c.WillRetain {
		flags |= 1 << 5
	}
	flags |= c.WillQos << 3
	if c.WillFlag {
		flags |= 1 << 2
	}
	if c.CleanStart {
		flags |= 1 << 1
	}
	if c.Reserved {
		flags |= 1
	}
	bufw.WriteByte(flags)
	// write keepalive
	bufw.WriteByte(byte(c.KeepAlive / 256))
	bufw.WriteByte(byte(c.KeepAlive % 256))
	// write properties
	if c.Version == V5 && c.Properties != nil {
		c.Properties.Pack(bufw)
	}
	// write payload
	bufw.Write(encodeString(c.ClientID))
	if c.WillFlag {

		bufw.Write(encodeString(c.WillTopic))
		bufw.Write(encodeString(c.WillMsg))
	}
	if c.UsernameFlag {
		bufw.Write(encodeString(c.Username))
	}
	if c.PasswordFlag {
		bufw.Write(encodeString(c.Password))
	}
	// write fix header
	c.FixHeader = &FixHeader{
		PacketType: CONNECT,
		RemainLen:  bufw.Len(),
	}
	if err := c.FixHeader.Pack(w); err != nil {
		return err
	}
	_, err := bufw.WriteTo(w)
	return err
}

// Unpack Connect Packet
func (c *Connect) Unpack(r io.Reader) error {
	var err error
	buf := make([]byte, c.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	c.Protocol = decodeString(bufr)
	if c.Version, err = bufr.ReadByte(); err != nil {
		return err
	}
	var flags byte
	if flags, err = bufr.ReadByte(); err != nil {
		return err
	}
	c.UsernameFlag = flags&0x80 > 0
	c.PasswordFlag = flags&0x40 > 0
	c.WillRetain = flags&0x20 > 0
	c.WillQos = flags & 0x18 >> 3
	c.WillFlag = flags&0x04 > 0
	c.CleanStart = flags&0x02 > 0
	c.Reserved = flags&0x01 > 0
	c.KeepAlive = readUint16(bufr)
	// unpack properties
	if c.Version == V5 {
		c.Properties = &Properties{}
		if err = c.Properties.Unpack(bufr); err != nil {
			return err
		}
	}
	// unpack payload
	c.ClientID = decodeString(bufr)
	if c.WillFlag {
		if c.Version == V5 {
			c.WillProperties = &Properties{}
			if err = c.Properties.Unpack(bufr); err != nil {
				return err
			}
		}
		c.WillTopic = decodeString(bufr)
		c.WillMsg = decodeString(bufr)
	}
	if c.UsernameFlag {
		c.Username = decodeString(bufr)
	}
	if c.PasswordFlag {
		c.Password = decodeString(bufr)
	}
	return nil
}
