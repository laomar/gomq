package store

import "github.com/laomar/gomq/store/topic"

type ram struct {
}

func NewRam() Store {
	return &ram{}
}

func (r *ram) NewTopicStore() (topic.Store, error) {
	return topic.NewRam(), nil
}
