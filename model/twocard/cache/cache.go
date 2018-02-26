package cache

import (
	"encoding/json"
	"fmt"
	mdtwo "playcards/model/twocard/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)

func TwoCardKey() string {
	return fmt.Sprintf(cache.KeyPrefix("TWOCARDMAP"))
}

func TwoCardSearchKey() string {
	return fmt.Sprintf(cache.KeyPrefix("TWOCARDSEARCH"))
}

func TwoCardSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func TwoCardLockKey(rid int32) string {
	return fmt.Sprintf("TWOCARDLOCK:%d", rid)
}

func SetGame(tc *mdtwo.Twocard) error {
	lockKey := TwoCardLockKey(tc.RoomID)
	key := TwoCardKey()
	searchKey := TwoCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := TwoCardSearchHKey(tc.Status, tc.RoomID)
			tc.SearchKey = searchHKey
			niuniu, _ := json.Marshal(tc)
			tx.HSet(key, tools.IntToString(tc.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, tc.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set two card failed", err)
	}
	return nil
}

func UpdateGame(tc *mdtwo.Twocard) error {
	lockKey := TwoCardLockKey(tc.RoomID)
	key := TwoCardKey()
	searchKey := TwoCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := TwoCardSearchHKey(tc.Status, tc.RoomID)
			lastKey := tc.SearchKey
			tx.HDel(searchKey, lastKey)
			tc.SearchKey = searchHKey
			niuniu, _ := json.Marshal(tc)
			tx.HSet(key, tools.IntToString(tc.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, tc.RoomID)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("update two card game failed", err)
	}

	return nil
}

func DeleteGame(tc *mdtwo.Twocard) error {
	lockKey := TwoCardLockKey(tc.RoomID)
	key := TwoCardKey()
	searchKey := TwoCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.IntToString(tc.RoomID))
			tx.HDel(searchKey, tc.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del two card game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdtwo.Twocard, error) {
	key := TwoCardKey()
	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get two card failed", err)
	}
	towcard := &mdtwo.Twocard{}
	if err := json.Unmarshal(val, towcard); err != nil {
		return nil, errors.Internal("get two card failed", err)
	}
	return towcard, nil
}

func GetAllGameByStatus(status int32) ([]*mdtwo.Twocard, error) {
	var curson uint64
	var ns []*mdtwo.Twocard
	var count int64
	count = 999
	key := TwoCardSearchKey()
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
						log.Err("get all two card key err rid:%s,err:%v", ridStr, err)
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
