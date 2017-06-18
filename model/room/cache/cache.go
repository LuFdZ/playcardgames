package cache

import (
	"encoding/json"
	mdr "playcards/model/room/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"

	"gopkg.in/redis.v5"
)

func RoomHKey(pwd string) string {
	return cache.KeyPrefix("ROOM:" + pwd)
}

func SetRoom(r *mdr.Room) error {
	return UpdateRoom(r)
}

func UpdateRoom(r *mdr.Room) error {
	key := RoomHKey(r.Password)

	f := func(tx *redis.Tx) error {
		orig, _ := tx.HGet(key, "pwd").Bytes()
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, string(orig))
			tx.HSet(key, "pwd", r.Password)
			b, _ := json.Marshal(r)
			tx.HSet(key, r.Password, string(b))
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set room error", err)
	}

	return nil
}

func DeleteRoom(pwd string) error {
	key := RoomHKey(pwd)
	f := func(tx *redis.Tx) error {
		orig, _ := tx.HGet(key, "pwd").Bytes()
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, string(orig))
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set room error", err)
	}

	return nil
}

func GetRoom(pwd string) (*mdr.Room, error) {

	key := RoomHKey(pwd)
	val, err := cache.KV().HGet(key, pwd).Bytes()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get room failed", err)
	}

	room := &mdr.Room{}
	if err := json.Unmarshal(val, room); err != nil {
		return nil, errors.Internal("get room failed", err)
	}

	return room, nil
}
