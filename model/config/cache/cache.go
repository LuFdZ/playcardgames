package cache

import (
	"bcr/utils/cache"
	"bcr/utils/errors"
	"bcr/utils/log"
	"encoding/json"
	"fmt"
	mdc "playcards/model/config/mod"

	"strconv"

	redis "gopkg.in/redis.v5"
)

func ConfigOpenHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ConfigOpen"))
}

func SetConfigOpens(cos []*mdc.ConfigOpen) error {
	err := DeleteConfigOpen()
	if err != nil {
		return err
	}
	var key string
	f := func(tx *redis.Tx) error {
		key = ConfigOpenHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, co := range cos {
				c, _ := json.Marshal(co)
				tx.HSet(key, strconv.Itoa(int(co.ItemID)), string(c))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set configopen list failed", err)
	}
	return nil
}

func SetConfigOpen(co *mdc.ConfigOpen) error {
	err := DeleteConfigOpen()
	if err != nil {
		return err
	}
	var key string
	f := func(tx *redis.Tx) error {
		key = ConfigOpenHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			c, _ := json.Marshal(co)
			tx.HSet(key, strconv.Itoa(int(co.ItemID)), string(c))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set configopen failed", err)
	}
	return nil
}

func GetConfigOpen(configOpenID string) (*mdc.ConfigOpen, error) {
	key := ConfigOpenHKey()
	val, err := cache.KV().HGet(key, configOpenID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get configopen failed", err)
	}

	co := &mdc.ConfigOpen{}
	if err := json.Unmarshal(val, co); err != nil {
		return nil, errors.Internal("get configopen failed", err)
	}
	return co, nil
}

func DeleteConfigOpen() error {
	key := ConfigOpenHKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete configopen error", err)
	}
	return nil
}

func GetAllConfigOpenKey() ([]string, error) {
	var curson uint64
	var rks []string
	var count int64
	count = 999
	for {
		scan := cache.KV().HScan(ConfigOpenHKey(), curson, "*", count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list configopen list failed", err)
		}

		curson = cur
		rks = append(rks, keys...)

		if curson == 0 {
			break
		}
	}
	return rks, nil
}

func GetAllConfigOpen(f func(*mdc.ConfigOpen) bool) []*mdc.ConfigOpen {
	var cos []*mdc.ConfigOpen
	keys, err := GetAllConfigOpenKey()
	if err != nil {
		log.Err("redis get all room err: %v", err)
	}
	for _, k := range keys {
		co, err := GetConfigOpen(k)
		if err != nil {
			log.Err("redis get room err: %v", err)
		}
		if co == nil {
			continue
		}
		if f != nil && !f(co) {
			continue
		}
		cos = append(cos, co)
	}
	return cos
}
