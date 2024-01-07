package subscription

import (
	"fmt"
	. "github.com/laomar/gomq/config"
	"github.com/syndtr/goleveldb/leveldb"
)

type diskStore struct {
	path string
	ram  *ramStore
}

func NewDiskStore() *diskStore {
	return &diskStore{
		path: Cfg.DataDir + "/subscription",
		ram:  NewRamStore(),
	}
}

func (s *diskStore) Init(cids []string) error {
	if len(cids) == 0 {
		return nil
	}
	db, err := leveldb.OpenFile(s.path, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	for _, cid := range cids {
		data, err := db.Get([]byte(cid), nil)
		if err != nil {
			return err
		}
		fmt.Println(data)
	}
	return nil
}

func (s *diskStore) Subscribe(cid string, sub *Subscription) error {
	return nil
}

func (r *diskStore) Unsubscribe(cid string, topic string) error {
	return nil
}

func (r *diskStore) UnsubscribeAll(cid string) error {
	return nil
}

func (r *diskStore) Close() error {
	return nil
}

func (r *diskStore) Print() {
}
