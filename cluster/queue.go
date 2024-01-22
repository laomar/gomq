package cluster

import (
	"container/list"
	"sync"
)

type queue struct {
	cond     *sync.Cond
	list     *list.List
	nextId   uint64
	nextRead *list.Element
}

func newQueue() *queue {
	return &queue{
		cond: sync.NewCond(&sync.Mutex{}),
		list: list.New(),
	}
}

func (q *queue) push(e *Event) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	e.Id = q.nextId
	q.nextId++
	elem := q.list.PushBack(e)
	if q.nextRead == nil {
		q.nextRead = elem
	}
	q.cond.Signal()
}

func (q *queue) pop(n int) []*Event {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for q.list.Len() == 0 || q.nextRead == nil {
		q.cond.Wait()
	}
	elem := q.nextRead
	es := make([]*Event, 0)
	for i := 0; i < n; i++ {
		es = append(es, elem.Value.(*Event))
		elem = elem.Next()
		if elem == nil {
			break
		}
	}
	q.nextRead = elem
	return es
}

func (q *queue) del(id uint64) {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	for elem := q.list.Front(); elem != nil; elem = elem.Next() {
		e := elem.Value.(*Event)
		if e.Id <= id {
			q.list.Remove(elem)
		}
		if e.Id == id {
			return
		}
	}
}
