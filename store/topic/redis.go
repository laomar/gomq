package topic

import (
	"context"
	"encoding/json"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/pkg/packets"
	goredis "github.com/redis/go-redis/v9"
)

type redis struct {
	ram    *ram
	db     goredis.UniversalClient
	prefix string
}

func NewRedis(db goredis.UniversalClient) *redis {
	redis := &redis{
		ram:    NewRam(),
		db:     db,
		prefix: prefix,
	}
	nodeName := config.Cfg.NodeName
	if nodeName != "" {
		redis.prefix += nodeName + ":"
	}
	return redis
}

func (r *redis) Init(cids ...string) error {
	if len(cids) == 0 {
		return nil
	}
	for _, cid := range cids {
		subs, err := r.db.HGetAll(context.Background(), r.prefix+cid).Result()
		if err != nil {
			return err
		}
		for _, s := range subs {
			sub := new(packets.Subscription)
			if err := json.Unmarshal([]byte(s), &sub); err != nil {
				return err
			}
			r.ram.Subscribe(cid, sub)
		}
	}
	return nil
}

func (r *redis) Subscribe(cid string, subs ...*packets.Subscription) (bool, error) {
	sets := make([]any, 0, len(subs)*2)
	for _, sub := range subs {
		jsub, _ := json.Marshal(sub)
		sets = append(sets, sub.Topic, string(jsub))
	}
	if _, err := r.db.HMSet(context.Background(), r.prefix+cid, sets...).Result(); err != nil {
		return false, err
	}
	return r.ram.Subscribe(cid, subs...)
}

func (r *redis) Unsubscribe(cid string, topics ...string) error {
	if _, err := r.db.HDel(context.Background(), r.prefix+cid, topics...).Result(); err != nil {
		return err
	}
	return r.ram.Unsubscribe(cid, topics...)
}

func (r *redis) UnsubscribeAll(cid string) error {
	if _, err := r.db.Del(context.Background(), r.prefix+cid).Result(); err != nil {
		return err
	}
	return r.ram.UnsubscribeAll(cid)
}

func (r *redis) Close() error {
	return r.db.Close()
}
