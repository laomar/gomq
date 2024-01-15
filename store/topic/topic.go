package topic

import (
	"github.com/laomar/gomq/pkg/packets"
)

const prefix = "topic:"

type Store interface {
	Init(...string) error
	Subscribe(string, ...*packets.Subscription) (bool, error)
	Unsubscribe(string, ...string) error
	UnsubscribeAll(string) error
	Close() error
}
