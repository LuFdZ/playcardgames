package cache

import (
	"encoding/json"
	"fmt"
	mdtow "playcards/model/towcard/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)

func TowCardKey() string {
	return fmt.Sprintf(cache.KeyPrefix("TOWCARDMAP"))
}

func TowCardSearchKey() string {
	return fmt.Sprintf(cache.KeyPrefix("TOWCARDSEARCH"))
}

func TowCardSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func TowCardLockKey(rid int32) string {
	return fmt.Sprintf("TOWCARDLOCK:%d", rid)
}

func SetGame(tc *mdtow.Towcard) error {
	lockKey := TowCardLockKey(tc.RoomID)
	key := TowCardKey()
	searchKey := TowCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := TowCardSearchHKey(tc.Status, tc.RoomID)
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
		return errors.Internal("set tow card failed", err)
	}
	return nil
}

func UpdateGame(tc *mdtow.Towcard) error {
	lockKey := TowCardLockKey(tc.RoomID)
	key := TowCardKey()
	searchKey := TowCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := TowCardSearchHKey(tc.Status, tc.RoomID)
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
		return errors.Internal("update tow card game failed", err)
	}

	return nil
}

func DeleteGame(tc *mdtow.Towcard) error {
	lockKey := TowCardLockKey(tc.RoomID)
	key := TowCardKey()
	searchKey := TowCardSearchKey()
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
		return errors.Internal("del tow card game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdtow.Towcard, error) {
	key := TowCardKey()
	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get tow card failed", err)
	}
	towcard := &mdtow.Towcard{}
	if err := json.Unmarshal(val, towcard); err != nil {
		return nil, errors.Internal("get tow card failed", err)
	}
	return towcard, nil
}

func GetAllGameByStatus(status int32) ([]*mdtow.Towcard, error) {
	var curson uint64
	var ns []*mdtow.Towcard
	var count int64
	count = 999
	key := TowCardSearchKey()
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
						log.Err("get all tow card key err rid:%s,err:%v", ridStr, err)
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
