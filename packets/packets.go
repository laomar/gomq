package packets

type FixHeader struct {
	PacketType byte
	Flags      byte
	RemainLen  int
}
