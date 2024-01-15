package store

import "github.com/laomar/gomq/store/topic"

type disk struct {
}

func NewDisk() Store {
	return &disk{}
}

func (d *disk) NewTopicStore() (topic.Store, error) {
	return topic.NewDisk()
}
