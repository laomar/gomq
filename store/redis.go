package store

import (
	"context"
	"github.com/laomar/gomq/config"
	"github.com/laomar/gomq/store/topic"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

type redis struct {
	db goredis.UniversalClient
}

func NewRedis() (Store, error) {
	cfg := config.Cfg.Store.Redis
	db := goredis.NewUniversalClient(&goredis.UniversalOptions{
		Addrs:       cfg.Addrs,
		Username:    cfg.User,
		Password:    cfg.Pwd,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	})
	_, err := db.Ping(context.Background()).Result()
	return &redis{
		db: db,
	}, err
}

func (r *redis) NewTopicStore() (topic.Store, error) {
	return topic.NewRedis(r.db), nil
}
