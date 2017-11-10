package cache

import (
	"playcards/utils/log"
	"encoding/json"
	"fmt"
	mdniu "playcards/model/niuniu/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"

	redis "gopkg.in/redis.v5"
)

func NiuniuHKey(rid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("NIUNIU:%d"), rid)
}

func NiuHKeySearch() string {
	return cache.KeyPrefix("NIUNIU:*")
}

func SetGame(n *mdniu.Niuniu, pwd string) error {
	key := NiuniuHKey(n.RoomID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			//tx.HSet(key, "userid", rid)
			niuniu, _ := json.Marshal(n)
			tx.HSet(key, "niuniu", string(niuniu))
			tx.HSet(key, "password", pwd)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set niuniu failed", err)
	}
	return nil
}

func UpdateGame(n *mdniu.Niuniu) error {
	key := NiuniuHKey(n.RoomID)
	ishas := cache.KV().Exists(key).Val()
	if !ishas {
		return nil
	}
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			niuniu, _ := json.Marshal(n)
			tx.HSet(key, "niuniu", string(niuniu))
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("update niuniu game failed", err)
	}

	return nil
}

func DeleteGame(rid int32) error {
	key := NiuniuHKey(rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del niuniu game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdniu.Niuniu, error) {
	key := NiuniuHKey(rid)
	val, err := cache.KV().HGet(key, "niuniu").Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get niuniu failed", err)
	}
	niuniu := &mdniu.Niuniu{}
	if err := json.Unmarshal(val, niuniu); err != nil {
		return nil, errors.Internal("get niuniu failed", err)
	}
	return niuniu, nil
}

func GetGameByKey(key string) (*mdniu.Niuniu, error) {
	//key := NiuniuHKey(rid)
	val, err := cache.KV().HGet(key, "niuniu").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get niuniu failed", err)
	}

	niuniu := &mdniu.Niuniu{}
	if err := json.Unmarshal(val, niuniu); err != nil {
		return nil, errors.Internal("get niuniu failed", err)
	}
	return niuniu, nil
}

func GetRoomPaawordRoomID(rid int32) string {
	key := NiuniuHKey(rid)
	pwd := cache.KV().HGet(key, "password").Val()
	return pwd
}


func GetAllNiuKey() ([]string, error) {
	var curson uint64
	var nks []string
	var count int64
	count = 999
	for {
		scan := cache.KV().Scan(curson, NiuHKeySearch(), count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list niuniu list failed", err)
		}

		curson = cur
		nks = append(nks, keys...)

		if curson == 0 {
			break
		}
	}
	return nks, nil
}

func GetAllNiu(f func(*mdniu.Niuniu) bool) []*mdniu.Niuniu {
	var niunius []*mdniu.Niuniu
	keys, err := GetAllNiuKey()
	if err != nil {
		log.Err("redis get all niuniu err: %v", err)
	}
	for _, k := range keys {
		niuniu, err := GetGameByKey(k)
		if err != nil {
			log.Err("redis get niuniu err: %v", err)
		}
		if niuniu == nil{
			continue
		}
		if f != nil && !f(niuniu) {
			continue
		}
		niunius = append(niunius, niuniu)
	}
	return niunius
}
