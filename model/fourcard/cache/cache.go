package cache

import (
	"encoding/json"
	"fmt"
	mdfour "playcards/model/fourcard/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)

func FourCardKey() string {
	return fmt.Sprintf(cache.KeyPrefix("FOURCARDMAP"))
}

func FourCardSearchKey() string {
	return fmt.Sprintf(cache.KeyPrefix("FOURCARDSEARCH"))
}

func FourCardSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func FourCardLockKey(rid int32) string {
	return fmt.Sprintf("FOURCARDLOCK:%d", rid)
}

func SetGame(fc *mdfour.Fourcard) error {
	lockKey := FourCardLockKey(fc.RoomID)
	key := FourCardKey()
	searchKey := FourCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := FourCardSearchHKey(fc.Status, fc.RoomID)
			fc.SearchKey = searchHKey
			niuniu, _ := json.Marshal(fc)
			tx.HSet(key, tools.IntToString(fc.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, fc.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set four card failed", err)
	}
	return nil
}

func UpdateGame(fc *mdfour.Fourcard) error {
	lockKey := FourCardLockKey(fc.RoomID)
	key := FourCardKey()
	searchKey := FourCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := FourCardSearchHKey(fc.Status, fc.RoomID)
			lastKey := fc.SearchKey
			tx.HDel(searchKey, lastKey)
			fc.SearchKey = searchHKey
			niuniu, _ := json.Marshal(fc)
			tx.HSet(key, tools.IntToString(fc.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, fc.RoomID)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("update four card game failed", err)
	}

	return nil
}

func DeleteGame(fc *mdfour.Fourcard) error {
	lockKey := FourCardLockKey(fc.RoomID)
	key := FourCardKey()
	searchKey := FourCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.IntToString(fc.RoomID))
			tx.HDel(searchKey, fc.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del four card game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdfour.Fourcard, error) {
	key := FourCardKey()
	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get four card failed", err)
	}
	niuniu := &mdfour.Fourcard{}
	if err := json.Unmarshal(val, niuniu); err != nil {
		return nil, errors.Internal("get four card failed", err)
	}
	return niuniu, nil
}

func GetAllGameByStatus(status int32) ([]*mdfour.Fourcard, error) {
	var curson uint64
	var ns []*mdfour.Fourcard
	var count int64
	count = 999
	key := FourCardSearchKey()
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
				statusValue, _ := tools.StringToInt(statusStr)
				if statusValue < status {
					ridStr := strings.Split(search[1], ":")[1]
					roomID, _ := tools.StringToInt(ridStr)
					niu, err := GetGame(roomID)
					if err != nil {
						log.Err("get all four card key err rid:%s,err:%v", ridStr, err)
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
