package topic

import (
	"github.com/laomar/gomq/pkg/packets"
	"strings"
	"sync"
)

type Ram struct {
	sync.RWMutex
	userTopic  *trie
	shareTopic *trie
}

func NewRam() *Ram {
	return &Ram{
		userTopic:  newTrie("user"),
		shareTopic: newTrie("share"),
	}
}

func (r *Ram) Init(cids ...string) error {
	return nil
}

func (r *Ram) Subscribe(cid string, subs ...*packets.Subscription) (bool, error) {
	defer r.Unlock()
	r.Lock()
	isExist := false
	for _, sub := range subs {
		if sub.ShareName != "" {
			isExist = r.shareTopic.subscribe(cid, sub)
		} else {
			isExist = r.userTopic.subscribe(cid, sub)
		}
	}
	return isExist, nil
}

func (r *Ram) Unsubscribe(cid string, topics ...string) error {
	defer r.Unlock()
	r.Lock()
	for _, topic := range topics {
		if strings.HasPrefix(topic, "$share") {
			r.shareTopic.unsubscribe(cid, topic)
		} else {
			r.userTopic.unsubscribe(cid, topic)
		}
	}
	return nil
}

func (r *Ram) UnsubscribeAll(cid string) error {
	defer r.Unlock()
	r.Lock()
	r.userTopic.unsubscribeAll(cid)
	r.shareTopic.unsubscribeAll(cid)
	return nil
}

func (r *Ram) Close() error {
	return nil
}

func (r *Ram) Print() {
	defer r.Unlock()
	r.Lock()
	r.userTopic.print()
	r.shareTopic.print()
}
