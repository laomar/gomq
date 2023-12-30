package subscription

import (
	"fmt"
	"strings"
	"sync"
)

type ramStore struct {
	sync.RWMutex
	topic    string
	parent   *ramStore
	children map[string]*ramStore
	subs     map[string]*Subscription
	shares   map[string]map[string]*Subscription
}

func NewRamStore() *ramStore {
	return &ramStore{
		children: make(map[string]*ramStore),
		subs:     make(map[string]*Subscription),
		shares:   make(map[string]map[string]*Subscription),
	}
}

func (r *ramStore) Init(cids []string) error {
	return nil
}

func (r *ramStore) print() {
	if r.parent == nil {
		fmt.Println()
	}
	fmt.Println(*r)
	for _, c := range r.children {
		c.print()
	}
}

func (r *ramStore) Subscribe(cid string, sub *Subscription) error {
	r.Lock()
	defer r.Unlock()
	topics := strings.Split(sub.Topic, "/")
	node := r
	for _, topic := range topics {
		if _, ok := node.children[topic]; !ok {
			child := NewRamStore()
			child.parent = node
			child.topic = topic
			node.children[topic] = child
		}
		node = node.children[topic]
	}
	if sub.ShareName != "" {
		if _, ok := node.shares[sub.ShareName]; !ok {
			node.shares[sub.ShareName] = make(map[string]*Subscription)
		}
		node.shares[sub.ShareName][cid] = sub
	} else {
		node.subs[cid] = sub
	}
	//r.print()
	return nil
}

func (r *ramStore) delete() {
	if r.parent == nil {
		return
	}
	if len(r.subs) == 0 && len(r.shares) == 0 && len(r.children) == 0 {
		delete(r.parent.children, r.topic)
		r.parent.delete()
	}
}

func (r *ramStore) Unsubscribe(cid string, topic string) error {
	r.Lock()
	defer r.Unlock()
	shareName, topicName := SplitTopic(topic)
	topics := strings.Split(topicName, "/")
	node := r
	for _, topic := range topics {
		if _, ok := node.children[topic]; !ok {
			return nil
		}
		node = node.children[topic]
	}
	if shareName != "" {
		if share, ok := node.shares[shareName]; ok {
			delete(share, cid)
			if len(share) == 0 {
				delete(node.shares, shareName)
			}
		}
	} else {
		delete(node.subs, cid)
	}
	node.delete()
	//r.print()
	return nil
}

func (r *ramStore) unsubscribeAll(cid string) {
	for _, c := range r.children {
		delete(c.subs, cid)
		for name, share := range c.shares {
			delete(share, cid)
			if len(share) == 0 {
				delete(c.shares, name)
			}
		}
		if len(c.children) > 0 {
			c.unsubscribeAll(cid)
		} else {
			c.delete()
		}

	}
}

func (r *ramStore) UnsubscribeAll(cid string) error {
	r.Lock()
	defer r.Unlock()
	r.unsubscribeAll(cid)
	r.print()
	return nil
}

func (r *ramStore) Close() error {
	return nil
}
