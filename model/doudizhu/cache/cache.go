package cache

import (
	"encoding/json"
	"fmt"
	mdddz "playcards/model/doudizhu/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
)
//
//func DoudizhuHKey(rid int32) string {
//	return fmt.Sprintf(cache.KeyPrefix("DOUDIZHU:%d"), rid)
//}
//
//func DoudizhuHKeySearch() string {
//	return cache.KeyPrefix("DOUDIZHU:*")
//}
//////////////////////////
func DoudizhuKey() string {
	return fmt.Sprintf(cache.KeyPrefix("DOUDIZHUMAP"))
}

func DoudizhuSearcKey() string {
	return fmt.Sprintf(cache.KeyPrefix("DOUDIZHUSEARCH"))
}

func DoudizhuSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func DoudizhuLockKey(rid int32) string {
	return fmt.Sprintf("DOUDIZHULOCK:%d", rid)
}


func SetGame(ddz *mdddz.Doudizhu) error {
	lockKey := DoudizhuLockKey(ddz.RoomID)
	key := DoudizhuKey()
	searchKey := DoudizhuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := DoudizhuSearchHKey(ddz.Status, ddz.RoomID)
			ddz.SearchKey = searchHKey
			d, _ := json.Marshal(ddz)
			tx.HSet(key, tools.IntToString(ddz.RoomID), string(d))
			tx.HSet(searchKey, searchHKey, ddz.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set doudizhu failed", err)
	}
	return nil
}

func UpdateGame(ddz *mdddz.Doudizhu) error {
	lockKey := DoudizhuLockKey(ddz.RoomID)
	key := DoudizhuKey()
	searchKey := DoudizhuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := DoudizhuSearchHKey(ddz.Status, ddz.RoomID)
			lastKey := ddz.SearchKey
			tx.HDel(searchKey, lastKey)
			ddz.SearchKey = searchHKey
			niuniu, _ := json.Marshal(ddz)
			tx.HSet(key, tools.IntToString(ddz.RoomID), string(niuniu))
			tx.HSet(searchKey, searchHKey, ddz.RoomID)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("update doudizhu game failed", err)
	}

	return nil
}

func DeleteGame(ddz *mdddz.Doudizhu) error {
	lockKey := DoudizhuLockKey(ddz.RoomID)
	key := DoudizhuKey()
	searchKey := DoudizhuSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.IntToString(ddz.RoomID))
			tx.HDel(searchKey, ddz.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del doudizhu game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdddz.Doudizhu, error) {
	key := DoudizhuKey()
	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
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

func GetAllDoudizhuByStatus(status int32) ([]*mdddz.Doudizhu, error) {
	var curson uint64
	var ddz []*mdddz.Doudizhu
	var count int64
	count = 999
	key := DoudizhuSearcKey()
	for {
		scan := cache.KV().HScan(key, curson, "*", count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list doudizhu list failed", err)
		}
		for i, searchDoudizhu := range keysValues {
			if i%2 == 0{
				search := strings.Split(searchDoudizhu, "-")
				statusStr := strings.Split(search[0], ":")[1]
				statusValue, _ := tools.StringToInt(statusStr)
				if statusValue < status {
					ridStr := strings.Split(search[1], ":")[1]
					roomID, _ := tools.StringToInt(ridStr)
					niu, err := GetGame(roomID)
					if err != nil {
						log.Err("GetAllDoudizhuKeyErr rid:%s,err:%v", ridStr, err)
					}
					ddz = append(ddz, niu)
				}
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return ddz, nil
}