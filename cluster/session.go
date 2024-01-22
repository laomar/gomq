package cluster

import (
	"container/list"
	"sync"
)

type session struct {
	id       string
	nodeName string
	nextId   uint64
	cache    *lru
	close    chan bool
}

type lru struct {
	list *list.List
	data map[uint64]struct{}
	size int
}

func (l *lru) push(id uint64) bool {
	if _, ok := l.data[id]; ok {
		return true
	}
	if len(l.data) == l.size {
		front := l.list.Front()
		l.list.Remove(front)
		delete(l.data, front.Value.(uint64))
	}
	l.data[id] = struct{}{}
	l.list.PushBack(id)
	return false
}

func newLru(size int) *lru {
	return &lru{
		list: list.New(),
		data: make(map[uint64]struct{}),
		size: size,
	}
}

type sessions struct {
	sync.RWMutex
	sessions map[string]*session
}

func (ss *sessions) get(nodeName string) *session {
	ss.RLock()
	defer ss.RUnlock()
	return ss.sessions[nodeName]
}

func (ss *sessions) del(nodeName string) {
	ss.Lock()
	defer ss.Unlock()
	if _, ok := ss.sessions[nodeName]; ok {
		delete(ss.sessions, nodeName)
	}
}

func (ss *sessions) set(nodeName, sid string) (restart bool, nextId uint64) {
	ss.Lock()
	defer ss.Unlock()
	if s, ok := ss.sessions[nodeName]; ok && s.id == sid {
		nextId = s.nextId
	} else {
		restart = true
		ss.sessions[nodeName] = &session{
			id:       sid,
			nodeName: nodeName,
			nextId:   0,
			cache:    newLru(100),
			close:    make(chan bool),
		}
	}
	return
}
