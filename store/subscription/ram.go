package subscription

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type ramStore struct {
	sync.RWMutex
	name     string
	parent   *ramStore
	child    sync.Map
	childnum int32
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
	fmt.Println(r.childnum)
	return nil
}

func (r *ramStore) print() {
	if r.parent == nil {
		fmt.Println()
	} else {
		fmt.Println(r.name, r.childnum, r.subs)
	}
	//for _, c := range r.children {
	//	c.Print()
	//}
	r.child.Range(func(_, c any) bool {
		c.(*ramStore).print()
		return true
	})
}

func (r *ramStore) Subscribe(cid string, sub *Subscription) error {
	defer r.Unlock()
	names := strings.Split(sub.Topic, "/")
	node := r
	for _, name := range names {
		if _, ok := node.child.Load(name); !ok {
			//if _, ok := node.children[name]; !ok {
			child := NewRamStore()
			child.parent = node
			child.name = name
			//node.children[name] = child
			node.child.Store(name, child)
			atomic.AddInt32(&node.childnum, 1)
		}
		//node = node.children[name]
		child, _ := node.child.Load(name)
		node = child.(*ramStore)
	}
	r.Lock()
	if sub.ShareName != "" {
		if _, ok := node.shares[sub.ShareName]; !ok {
			node.shares[sub.ShareName] = make(map[string]*Subscription)
		}
		node.shares[sub.ShareName][cid] = sub
	} else {
		node.subs[cid] = sub
	}
	return nil
}

func (r *ramStore) delete() {
	if r.parent == nil || r.childnum > 0 {
		return
	}
	if len(r.subs) == 0 && len(r.shares) == 0 {
		//delete(r.parent.children, r.name)
		r.parent.child.Delete(r.name)
		if atomic.AddInt32(&r.parent.childnum, -1) == 0 {
			r.parent.delete()
		}
	}
}

func (r *ramStore) Unsubscribe(cid string, topic string) error {
	defer r.Unlock()
	shareName, topicName := SplitTopic(topic)
	names := strings.Split(topicName, "/")
	node := r
	for _, name := range names {
		if _, ok := node.child.Load(name); !ok {
			//if _, ok := node.children[name]; !ok {
			return nil
		}
		//node = node.children[name]
		child, _ := node.child.Load(name)
		node = child.(*ramStore)
	}
	r.Lock()
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
	r.print()
	return nil
}

func (r *ramStore) unsubscribeAll(cid string) {
	//for _, c := range r.children {
	//	delete(c.subs, cid)
	//	for name, share := range c.shares {
	//		delete(share, cid)
	//		if len(share) == 0 {
	//			delete(c.shares, name)
	//		}
	//	}
	//	if len(c.children) > 0 {
	//		c.unsubscribeAll(cid)
	//	} else {
	//		c.delete()
	//	}
	//}
	r.child.Range(func(_, c any) bool {
		child := c.(*ramStore)
		delete(child.subs, cid)
		for name, share := range child.shares {
			delete(share, cid)
			if len(share) == 0 {
				delete(child.shares, name)
			}
		}
		if child.childnum > 0 {
			child.unsubscribeAll(cid)
		} else {
			child.delete()
		}
		return true
	})
}

func (r *ramStore) UnsubscribeAll(cid string) error {
	r.Lock()
	defer r.Unlock()
	r.unsubscribeAll(cid)
	return nil
}

func (r *ramStore) Close() error {
	return nil
}
