package packets

type Publish struct {
	FixHeader
	Dup    bool
	Qos    byte
	Retain bool
}
