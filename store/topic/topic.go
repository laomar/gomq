package topic

import (
	"github.com/laomar/gomq/pkg/packets"
	"strings"
)

type Store interface {
	Init(cids []string) error
	Subscribe(cid string, subs ...*packets.Subscription) error
	Unsubscribe(cid string, topic ...string) error
	UnsubscribeAll(cid string) error
	Close() error
}

func Split(topic string) (string, string) {
	if strings.HasPrefix(topic, "$share") {
		shares := strings.SplitN(topic, "/", 3)
		length := len(shares)
		switch {
		case length > 2:
			return shares[1], shares[2]
		case length == 2:
			return shares[1], ""
		default:
			return "", ""
		}
	}
	return "", topic
}
