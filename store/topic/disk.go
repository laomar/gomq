package topic

import (
	"encoding/json"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/pkg/packets"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type disk struct {
	ram *ram
	db  *leveldb.DB
}

func NewDisk() (*disk, error) {
	db, err := leveldb.OpenFile(config.Cfg.DataDir+"/topic", nil)
	if err != nil {
		return nil, err
	}
	return &disk{
		db:  db,
		ram: NewRam(),
	}, nil
}

func (d *disk) Init(cids ...string) error {
	if len(cids) == 0 {
		return nil
	}
	for _, cid := range cids {
		iter := d.db.NewIterator(util.BytesPrefix([]byte(prefix+cid)), nil)
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

func (d *disk) Subscribe(cid string, subs ...*packets.Subscription) (bool, error) {
	batch := new(leveldb.Batch)
	for _, sub := range subs {
		jsub, _ := json.Marshal(sub)
		batch.Put([]byte(prefix+cid+":"+sub.Topic), jsub)
	}
	if err := d.db.Write(batch, nil); err != nil {
		return false, err
	}
	return d.ram.Subscribe(cid, subs...)
}

func (d *disk) Unsubscribe(cid string, topics ...string) error {
	batch := new(leveldb.Batch)
	for _, topic := range topics {
		batch.Delete([]byte(prefix + cid + ":" + topic))
	}
	if err := d.db.Write(batch, nil); err != nil {
		return err
	}
	return d.ram.Unsubscribe(cid, topics...)
}

func (d *disk) UnsubscribeAll(cid string) error {
	iter := d.db.NewIterator(util.BytesPrefix([]byte(prefix+cid)), nil)
	batch := new(leveldb.Batch)
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	iter.Release()
	if err := d.db.Write(batch, nil); err != nil {
		return err
	}
	return d.ram.UnsubscribeAll(cid)
}

func (d *disk) Close() error {
	return d.db.Close()
}
