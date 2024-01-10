package topic

import (
	"github.com/laomar/gomq/pkg/packets"
	"strings"
	"sync"
)

type ram struct {
	sync.RWMutex
	userTopic  *trie
	shareTopic *trie
}

func NewRam() *ram {
	return &ram{
		userTopic:  newTrie("user"),
		shareTopic: newTrie("share"),
	}
}

func (r *ram) Init(cids []string) error {
	return nil
}

func (r *ram) Subscribe(cid string, subs ...*packets.Subscription) error {
	defer r.Unlock()
	r.Lock()
	for _, sub := range subs {
		if strings.HasPrefix(sub.Topic, "$share") {
			r.shareTopic.subscribe(cid, sub)
		} else {
			r.userTopic.subscribe(cid, sub)
		}
	}
	//r.userTopic.print()
	//r.shareTopic.print()
	return nil
}

func (r *ram) Unsubscribe(cid string, topics ...string) error {
	defer r.Unlock()
	r.Lock()
	for _, topic := range topics {
		if strings.HasPrefix(topic, "$share") {
			r.shareTopic.unsubscribe(cid, topic)
		} else {
			r.userTopic.unsubscribe(cid, topic)
		}
	}
	//r.userTopic.print()
	//r.shareTopic.print()
	return nil
}

func (r *ram) UnsubscribeAll(cid string) error {
	defer r.Unlock()
	r.Lock()
	r.userTopic.unsubscribeAll(cid)
	//r.userTopic.print()
	r.shareTopic.unsubscribeAll(cid)
	//r.shareTopic.print()
	return nil
}

func (r *ram) Close() error {
	return nil
}
