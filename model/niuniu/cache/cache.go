package cache

import (
	"encoding/json"
	"fmt"
	mdniu "playcards/model/niuniu/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)

func NiuniuKey() string {
	return fmt.Sprintf(cache.KeyPrefix("NIUNIUMAP"))
}

func NiuniuSearcKey() string {
	return fmt.Sprintf(cache.KeyPrefix("NIUNIUSEARCH"))
}

func NiuniuSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func ThirteenLockKey(rid int32) string {
	return fmt.Sprintf("THIRTEENLOCK:%d", rid)
}

func SetGame(n *mdniu.Niuniu) error {
	lockKey := ThirteenLockKey(n.RoomID)
	key := NiuniuKey()
	searchKey := NiuniuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := NiuniuSearchHKey(n.Status, n.RoomID)
			n.SearchKey = searchHKey
			niuniu, _ := json.Marshal(n)
			tx.HSet(key, tools.String2int(n.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, n.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set niuniu failed", err)
	}
	return nil
}

func UpdateGame(n *mdniu.Niuniu) error {
	lockKey := ThirteenLockKey(n.RoomID)
	key := NiuniuKey()
	searchKey := NiuniuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := NiuniuSearchHKey(n.Status, n.RoomID)
			lastKey := n.SearchKey
			tx.HDel(searchKey, lastKey)
			n.SearchKey = searchHKey
			niuniu, _ := json.Marshal(n)
			tx.HSet(key, tools.String2int(n.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, n.RoomID)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("update niuniu game failed", err)
	}

	return nil
}

func DeleteGame(n *mdniu.Niuniu) error {
	lockKey := ThirteenLockKey(n.RoomID)
	key := NiuniuKey()
	searchKey := NiuniuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.String2int(n.RoomID))
			tx.HDel(searchKey, n.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del niuniu game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdniu.Niuniu, error) {
	key := NiuniuKey()
	val, err := cache.KV().HGet(key, tools.String2int(rid)).Bytes()
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

func GetAllNiuniuByStatus(status int32) ([]*mdniu.Niuniu, error) {
	var curson uint64
	var ns []*mdniu.Niuniu
	var count int64
	count = 999
	key := NiuniuSearcKey()
	for {
		scan := cache.KV().HScan(key, curson, "*", count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list room list failed", err)
		}
		for i, searchNiuniu := range keysValues {
			if i%2 == 0 {
				search := strings.Split(searchNiuniu, "-")
				statusStr := strings.Split(search[0], ":")[1]
				statusValue, _ := tools.Int2String(statusStr)
				if statusValue < status {
					ridStr := strings.Split(search[1], ":")[1]
					roomID, _ := tools.Int2String(ridStr)
					niu, err := GetGame(roomID)
					if err != nil {
						log.Err("GetAllDoudizhuKeyErr rid:%s,err:%v", ridStr, err)
					}
					ns = append(ns, niu)
				}
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return ns, nil
}
