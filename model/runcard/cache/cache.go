package cache

import (
	"encoding/json"
	"fmt"
	mdrun "playcards/model/runcard/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)

func RunCardKey() string {
	return fmt.Sprintf(cache.KeyPrefix("RUNCARDMAP"))
}

func RunCardSearchKey() string {
	return fmt.Sprintf(cache.KeyPrefix("RUNCARDSEARCH"))
}

func RunCardSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func RunCardLockKey(rid int32) string {
	return fmt.Sprintf("RUNCARDLOCK:%d", rid)
}

func SetGame(rc *mdrun.Runcard) error {
	lockKey := RunCardLockKey(rc.RoomID)
	key := RunCardKey()
	searchKey := RunCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := RunCardSearchHKey(rc.Status, rc.RoomID)
			rc.SearchKey = searchHKey
			game, _ := json.Marshal(rc)
			tx.HSet(key, tools.IntToString(rc.RoomID), string(game))
			tx.HSet(searchKey, searchHKey, rc.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set run card failed", err)
	}
	return nil
}

func UpdateGame(rc *mdrun.Runcard) error {
	lockKey := RunCardLockKey(rc.RoomID)
	key := RunCardKey()
	searchKey := RunCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := RunCardSearchHKey(rc.Status, rc.RoomID)
			lastKey := rc.SearchKey
			tx.HDel(searchKey, lastKey)
			rc.SearchKey = searchHKey
			niuniu, _ := json.Marshal(rc)
			tx.HSet(key, tools.IntToString(rc.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, rc.RoomID)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("update run card game failed", err)
	}

	return nil
}

func DeleteGame(rc *mdrun.Runcard) error {
	lockKey := RunCardLockKey(rc.RoomID)
	key := RunCardKey()
	searchKey := RunCardSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.IntToString(rc.RoomID))
			tx.HDel(searchKey, rc.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del run card game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdrun.Runcard, error) {
	key := RunCardKey()
	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get run card failed", err)
	}
	towcard := &mdrun.Runcard{}
	if err := json.Unmarshal(val, towcard); err != nil {
		return nil, errors.Internal("get run card failed", err)
	}
	return towcard, nil
}

func GetAllGameByStatus(status int32) ([]*mdrun.Runcard, error) {
	var curson uint64
	var ns []*mdrun.Runcard
	var count int64
	count = 999
	key := RunCardSearchKey()
	for {
		scan := cache.KV().HScan(key, curson, "*", count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("room list failed", err)
		}
		for i, searchNiuniu := range keysValues {
			if i%2 == 0 {
				search := strings.Split(searchNiuniu, "-")
				statusStr := strings.Split(search[0], ":")[1]
				statusValue, _ := tools.StringToInt(statusStr)
				if statusValue == status {
					ridStr := strings.Split(search[1], ":")[1]
					roomID, _ := tools.StringToInt(ridStr)
					niu, err := GetGame(roomID)
					if err != nil {
						log.Err("get all run card key err rid:%s,err:%v", ridStr, err)
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
