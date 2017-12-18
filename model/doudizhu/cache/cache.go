package cache

import (
	"encoding/json"
	"fmt"
	mdddz "playcards/model/doudizhu/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
)

func DoudizhuHKey(rid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("DOUDIZHU:%d"), rid)
}

func DoudizhuHKeySearch() string {
	return cache.KeyPrefix("DOUDIZHU:*")
}

func SetGame(ddz *mdddz.Doudizhu, pwd string) error {
	key := DoudizhuHKey(ddz.RoomID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			//tx.HSet(key, "userid", rid)
			niuniu, _ := json.Marshal(ddz)
			tx.HSet(key, "doudizhu", string(niuniu))
			tx.HSet(key, "password", pwd)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set doudizhu failed", err)
	}
	return nil
}

func UpdateGame(ddz *mdddz.Doudizhu) error {
	key := DoudizhuHKey(ddz.RoomID)
	ishas := cache.KV().Exists(key).Val()
	if !ishas {
		return nil
	}
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			niuniu, _ := json.Marshal(ddz)
			tx.HSet(key, "doudizhu", string(niuniu))
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("update doudizhu game failed", err)
	}

	return nil
}

func DeleteGame(rid int32) error {
	key := DoudizhuHKey(rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del doudizhu game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdddz.Doudizhu, error) {
	key := DoudizhuHKey(rid)
	val, err := cache.KV().HGet(key, "doudizhu").Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get doudizhu failed", err)
	}
	doudizhu := &mdddz.Doudizhu{}
	if err := json.Unmarshal(val, doudizhu); err != nil {
		return nil, errors.Internal("get doudizhu failed", err)
	}
	return doudizhu, nil
}

func GetRoomPaawordRoomID(rid int32) string {
	key := DoudizhuHKey(rid)
	pwd := cache.KV().HGet(key, "password").Val()
	return pwd
}

func GetGameByKey(key string) (*mdddz.Doudizhu, error) {
	//key := NiuniuHKey(rid)
	val, err := cache.KV().HGet(key, "doudizhu").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get doudizhu failed", err)
	}

	doudizhu := &mdddz.Doudizhu{}
	if err := json.Unmarshal(val, doudizhu); err != nil {
		return nil, errors.Internal("get doudizhu failed", err)
	}
	return doudizhu, nil
}

func GetAllDDZKey() ([]string, error) {
	var curson uint64
	var nks []string
	var count int64
	count = 999
	for {
		scan := cache.KV().Scan(curson, DoudizhuHKeySearch(), count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list doudizhu list failed", err)
		}

		curson = cur
		nks = append(nks, keys...)

		if curson == 0 {
			break
		}
	}
	return nks, nil
}

func GetAllDDZ(f func(ddz *mdddz.Doudizhu) bool) []*mdddz.Doudizhu {
	var ddzs []*mdddz.Doudizhu
	keys, err := GetAllDDZKey()
	if err != nil {
		log.Err("redis get all doudizhu err: %v", err)
	}
	for _, k := range keys {
		ddz, err := GetGameByKey(k)
		if err != nil {
			log.Err("redis get doudizhu err: %v", err)
		}
		if ddz == nil {
			continue
		}
		if f != nil && !f(ddz) {
			continue
		}
		ddzs = append(ddzs, ddz)
	}
	return ddzs
}
