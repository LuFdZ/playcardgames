package cache

import (
	"encoding/json"
	"fmt"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	"gopkg.in/redis.v5"
	"strings"
	"playcards/utils/tools"
)

func ThirteenKey() string {
	return fmt.Sprintf(cache.KeyPrefix("THIRTEENMAP"))
}

func ThirteenSearcKey() string {
	return fmt.Sprintf(cache.KeyPrefix("THIRTEENSEARCH"))
}

func ThirteenSearchHKey(status int32, rid int32) string {
	return fmt.Sprintf("status:%d-rid:%d-", status, rid)
}

func ThirteenLockKey(rid int32) string {
	return fmt.Sprintf("THIRTEENLOCK:%d", rid)
}

func SetGame(t *mdt.Thirteen) error {
	lockKey := ThirteenLockKey(t.RoomID)
	key := ThirteenKey()
	searchKey := ThirteenSearcKey()
	f := func(tx *redis.Tx) error {
		searchHKey := ThirteenSearchHKey(t.Status, t.RoomID)
		tx.Pipelined(func(p *redis.Pipeline) error {
			t.SearchKey = searchHKey
			thirteen, _ := json.Marshal(t)
			tx.HSet(key, tools.String2int(t.RoomID), string(thirteen))
			tx.HSet(searchKey, searchHKey, t.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set thirteen user failed", err)
	}
	return nil
}

func UpdateGame(t *mdt.Thirteen) error {
	lockKey := ThirteenLockKey(t.RoomID)
	key := ThirteenKey()
	searchKey := ThirteenSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			searchHKey := ThirteenSearchHKey(t.Status, t.RoomID)
			lastKey := t.SearchKey
			tx.HDel(searchKey, lastKey)
			t.SearchKey = searchHKey
			b, _ := json.Marshal(t)
			tx.HSet(key, tools.String2int(t.RoomID), string(b))
			tx.HSet(searchKey, searchHKey, t.RoomID)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set thirteen user failed", err)
	}
	return nil
}

func DeleteGame(t *mdt.Thirteen) error {
	lockKey := ThirteenLockKey(t.RoomID)
	key := ThirteenKey()
	searchKey := ThirteenSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key,tools.String2int(t.RoomID))
			tx.HDel(searchKey, t.SearchKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("del thirteen game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdt.Thirteen, error) {
	key := ThirteenKey()
	val, err := cache.KV().HGet(key, tools.String2int(rid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get thirteen failed", err)
	}

	thirteen := &mdt.Thirteen{}
	if err := json.Unmarshal(val, thirteen); err != nil {
		return nil, errors.Internal("get thirteen failed", err)
	}

	return thirteen, nil
}

func GetAllThirteenByStatus(status int32) []*mdt.Thirteen {
	var thirteens []*mdt.Thirteen
	match := fmt.Sprintf("*status:%d-*", status)
	thirteens, err := GetMatchThirteen(match)
	if err != nil {
		log.Err("GetAllRoomByStatus:%v", err)
	}
	return thirteens
}

func GetMatchThirteen(match string) ([]*mdt.Thirteen, error) {
	var curson uint64
	var ts []*mdt.Thirteen
	var count int64
	count = 999
	key := ThirteenSearcKey()
	for {
		scan := cache.KV().HScan(key, curson, match, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list room list failed", err)
		}
		for i, searchThirteen := range keysValues {
			if i%2==0{
				rid := strings.Split(searchThirteen,"-")[1]
				ridStr := strings.Split(rid,":")[1]
				roomID,_:= tools.Int2String(ridStr)
				room ,err:= GetGame(roomID)
				if err!=nil{
					log.Err("GetAllThirteenKeyErr match:%s,err:%v",match,err)
				}
				ts = append(ts, room)
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return ts, nil
}

