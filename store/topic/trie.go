package topic

import (
	"fmt"
	"github.com/laomar/gomq/pkg/packets"
	"strings"
)

type trie struct {
	name     string
	parent   *trie
	children map[string]*trie
	subs     map[string]*packets.Subscription
}

func newTrie(name string) *trie {
	return &trie{
		name:     name,
		children: make(map[string]*trie),
		subs:     make(map[string]*packets.Subscription),
	}
}

func (t *trie) subscribe(cid string, sub *packets.Subscription) bool {
	isExist := true
	names := strings.Split(sub.Topic, "/")
	node := t
	for _, name := range names {
		if name == "$share" {
			continue
		}
		if _, ok := node.children[name]; !ok {
			child := newTrie(name)
			child.parent = node
			node.children[name] = child
			isExist = false
		}
		node = node.children[name]
	}
	if node.subs[cid] == nil {
		isExist = false
	}
	node.subs[cid] = sub
	return isExist
}

func (t *trie) print() {
	fmt.Println(t.name, t.subs)
	for _, c := range t.children {
		c.print()
	}
}

func (t *trie) unsubscribe(cid string, topic string) {
	names := strings.Split(topic, "/")
	node := t
	for _, name := range names {
		if name == "$share" {
			continue
		}
		if _, ok := node.children[name]; !ok {
			return
		}
		node = node.children[name]
	}
	delete(node.subs, cid)
	node.delete()
}

func (t *trie) delete() {
	if t.parent == nil || len(t.children) > 0 || len(t.subs) > 0 {
		return
	}
	delete(t.parent.children, t.name)
	t.parent.delete()
}

func (t *trie) unsubscribeAll(cid string) {
	for _, c := range t.children {
		delete(c.subs, cid)
		if len(c.children) > 0 {
			c.unsubscribeAll(cid)
		} else {
			c.delete()
		}
	}
}
