package packets

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Connect Packet
type Connect struct {
	FixHeader *FixHeader
	// Variable Header
	ProtocolName    string
	ProtocolVersion byte
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
func (c *Connect) pack(w io.Writer) error {
	var err error
	var buf bytes.Buffer
	buf.WriteByte(0x00)
	buf.WriteByte(0x04)
	buf.Write([]byte(c.ProtocolName))
	buf.WriteByte(c.ProtocolVersion)
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
	buf.WriteByte(flags)
	// write keepalive
	buf.WriteByte(byte(c.KeepAlive / 256))
	buf.WriteByte(byte(c.KeepAlive % 256))
	// write properties

	// write payload
	buf.Write(encodeString(c.ClientID))
	if c.WillFlag {

		buf.Write(encodeString(c.WillTopic))
		buf.Write(encodeString(c.WillMsg))
	}
	if c.UsernameFlag {
		buf.Write(encodeString(c.Username))
	}
	if c.PasswordFlag {
		buf.Write(encodeString(c.Password))
	}
	// write fix header
	c.FixHeader.PacketType = CONNECT
	c.FixHeader.RemainLen = buf.Len()
	err = c.FixHeader.pack(w)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

// Unpack Connect Packet
func (c *Connect) unpack(r io.Reader) error {
	var err error
	buf := make([]byte, c.FixHeader.RemainLen)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	bufr := bytes.NewBuffer(buf)
	c.ProtocolName = decodeString(bufr)
	if c.ProtocolVersion, err = bufr.ReadByte(); err != nil {
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
	c.KeepAlive = binary.BigEndian.Uint16(bufr.Next(2))
	// unpack properties
	if c.ProtocolVersion == V5 {
		c.Properties = &Properties{}
		if err = c.Properties.unpack(bufr); err != nil {
			return err
		}
	}
	// unpack payload
	c.ClientID = decodeString(bufr)
	if c.WillFlag {
		if c.ProtocolVersion == V5 {
			c.WillProperties = &Properties{}
			if err = c.Properties.unpack(bufr); err != nil {
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
