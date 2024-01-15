package store

import (
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/store/topic"
)

type Store interface {
	NewTopicStore() (topic.Store, error)
}

func NewStore() (Store, error) {
	var err error
	var se Store
	switch config.Cfg.Store.Type {
	case "disk":
		se = NewDisk()
	case "redis":
		se, err = NewRedis()
	default:
		se = NewRam()
	}
	return se, err
}
