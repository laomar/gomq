package server

type Plugin interface {
	Name() string
	Load() error
	Unload() error
}

type NewPlugin func() (Plugin, error)
