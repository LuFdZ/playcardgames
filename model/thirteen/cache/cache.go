package cache

import (
	"encoding/json"
	"fmt"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"strconv"

	"gopkg.in/redis.v5"
)

func ThirteenHKey(rid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("THIRTEEN:%s"), rid)
}

func SetGame(t *mdt.Thirteen, playernum int32, pwd string) error {
	key := ThirteenHKey(t.RoomID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			//tx.HSet(key, "userid", rid)
			thirteen, _ := json.Marshal(t)
			tx.HSet(key, "thirteen", string(thirteen))
			tx.HSet(key, "gameid", t.GameID)
			tx.HSet(key, "playernum", playernum)
			tx.HSet(key, "playernow", 0)
			tx.HSet(key, "password", pwd)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set thirteen user failed", err)
	}
	return nil
}

func SetGameUser(rid int32, uid int32) error {
	key := ThirteenHKey(rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, string(uid), 1)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set game user failed", err)
	}

	return nil
}

func UpdateGameUser(t *mdt.Thirteen, uid int32, playernow int32) error {
	key := ThirteenHKey(t.RoomID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			thirteen, _ := json.Marshal(t)
			tx.HSet(key, "thirteen", string(thirteen))
			tx.HSet(key, "playernow", playernow)
			tx.HSet(key, string(uid), 2)
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("update niuniu game user failed", err)
	}

	return nil
}

func DeleteGame(rid int32) error {
	key := ThirteenHKey(rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {

			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del thirteen game redis error", err)
	}
	return nil
}

func GetGame(rid int32) (*mdt.Thirteen, error) {
	key := ThirteenHKey(rid)
	val, err := cache.KV().HGet(key, "thirteen").Bytes()
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

func GetGamePlayerNumRoomID(rid int32) int32 {
	key := ThirteenHKey(rid)
	num := cache.KV().HGet(key, "playernum").Val()
	if len(num) > 0 {
		result, _ := strconv.Atoi(num)
		return int32(result)
	}
	return 0
}

func GetGamePlayerNowRoomID(rid int32) int32 {
	key := ThirteenHKey(rid)
	num := cache.KV().HGet(key, "playernow").Val()
	if len(num) > 0 {
		result, _ := strconv.Atoi(num)
		return int32(result)
	}
	return 0
}

func GetRoomPaawordRoomID(rid int32) string {
	key := ThirteenHKey(rid)
	pwd := cache.KV().HGet(key, "password").Val()
	return pwd
}

func IsGamePlayerReady(rid int32, uid int32) int32 {
	key := ThirteenHKey(rid)
	num := cache.KV().HGet(key, string(uid)).Val()
	if len(num) > 0 {
		result, _ := strconv.Atoi(num)
		return int32(result)
	}
	return 0
}
