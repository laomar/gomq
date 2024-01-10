package topic

import (
	"encoding/json"
	"fmt"
	. "github.com/laomar/gomq/config"
	"github.com/laomar/gomq/pkg/packets"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const prefix = "topic:"

type disk struct {
	path string
	ram  *ram
}

func NewDisk() *disk {
	return &disk{
		path: Cfg.DataDir + "/topic",
		ram:  NewRam(),
	}
}

func (d *disk) Init(cids []string) error {
	if len(cids) == 0 {
		return nil
	}
	db, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Println(string(iter.Key()))
	}
	iter.Release()

	for _, cid := range cids {
		iter := db.NewIterator(util.BytesPrefix([]byte(prefix+cid)), nil)
		for iter.Next() {
			sub := new(packets.Subscription)
			if err := json.Unmarshal(iter.Value(), &sub); err != nil {
				return err
			}
			d.ram.Subscribe(cid, sub)
		}
		iter.Release()
	}
	return nil
}

func (d *disk) Subscribe(cid string, subs ...*packets.Subscription) error {
	db, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	for _, sub := range subs {
		jsub, _ := json.Marshal(sub)
		batch.Put([]byte(prefix+cid+":"+sub.Topic), jsub)
	}
	if err := db.Write(batch, nil); err != nil {
		return err
	}
	db.Close()
	return d.ram.Subscribe(cid, subs...)
}

func (d *disk) Unsubscribe(cid string, topics ...string) error {
	db, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	for _, topic := range topics {
		batch.Delete([]byte(prefix + cid + ":" + topic))
	}
	if err := db.Write(batch, nil); err != nil {
		return err
	}
	db.Close()
	return d.ram.Unsubscribe(cid, topics...)
}

func (d *disk) UnsubscribeAll(cid string) error {
	db, err := leveldb.OpenFile(d.path, nil)
	if err != nil {
		return err
	}
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix+cid)), nil)
	batch := new(leveldb.Batch)
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	iter.Release()
	if err := db.Write(batch, nil); err != nil {
		return err
	}
	db.Close()
	return d.ram.UnsubscribeAll(cid)
}

func (d *disk) Close() error {
	return nil
}
