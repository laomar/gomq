package subscription

import "strings"

type Subscription struct {
	Topic             string
	ShareName         string
	RetainHandling    byte
	RetainAsPublished bool
	NoLocal           bool
	Qos               byte
	SubID             uint32
}

type Store interface {
	Init(cids []string) error
	Subscribe(cid string, sub *Subscription) error
	Unsubscribe(cid string, topic string) error
	UnsubscribeAll(cid string) error
	Close() error
}

func SplitTopic(topic string) (string, string) {
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
