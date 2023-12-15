package packets

import (
	"io"
)

// FixHeader Struct
type FixHeader struct {
	PacketType byte
	Flags      byte
	Dup        bool
	Qos        byte
	Retain     bool
	RemainLen  int
}

// Pack Fix Header
func (fh *FixHeader) Pack(w io.Writer) error {
	if fh.PacketType == PUBLISH {
		if fh.Dup {
			fh.Flags |= 1 << 3
		}
		fh.Flags |= fh.Qos << 1
		if fh.Retain {
			fh.Flags |= 1
		}
	}
	b := make([]byte, 1)
	b[0] = fh.PacketType<<4 | fh.Flags
	b = append(b, encodeLength(fh.RemainLen)...)
	_, err := w.Write(b)
	return err
}

// Unpack Fix Header
func (fh *FixHeader) Unpack(r io.Reader) error {
	var err error
	b := make([]byte, 1)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	fh.PacketType = b[0] >> 4
	fh.Flags = b[0] & 0x0F
	if fh.PacketType == PUBLISH {
		fh.Dup = fh.Flags>>3 > 0
		fh.Qos = (fh.Flags >> 1) & 0x03
		fh.Retain = fh.Flags&0x01 > 0
	}
	if fh.RemainLen, err = decodeLength(r); err != nil {
		return err
	}
	return nil
}
