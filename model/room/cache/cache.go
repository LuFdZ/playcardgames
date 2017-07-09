package cache

import (
	"encoding/json"
	"fmt"
	mdr "playcards/model/room/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"

	"gopkg.in/redis.v5"
)

func RoomHKey(pwd string) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOM:%d"), pwd)
}

func UserHKey(uid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMUSER:%n"), uid)
}

func SetRoom(r *mdr.Room) error {
	key := RoomHKey(r.Password)

	f := func(tx *redis.Tx) error {
		roomInfo := tx.HGetAll(key).Val()
		if roomInfo != nil {
			return errors.Internal("room password has exist", nil)
		}

		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(r)

			tx.HSet(key, "password", r.Password)
			tx.HSet(key, "roomid", r.RoomID)
			tx.HSet(key, "room", b)
			return nil
		})

		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set room failed", err)
	}

	return nil
}

func UpdateRoom(r *mdr.Room) error {
	key := RoomHKey(r.Password)

	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(r)
			tx.HSet(key, "room", string(b))
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set room failed", err)
	}

	return nil
}

func DeleteRoom(password string) error {
	key := RoomHKey(password)

	f := func(tx *redis.Tx) error {
		orig, _ := tx.HGet(key, "password").Bytes()
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

func SetRoomUser(rid int32, password string, uid int32) error {
	key := UserHKey(uid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "userid", rid)
			tx.HSet(key, "roomid", rid)
			tx.HSet(key, "password", password)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set roomuser failed", err)
	}
	return nil
}

func GetRoomPasswordByUserID(uid int32) string {
	key := UserHKey(uid)
	pwd := cache.KV().HGet(key, "roomid").String()
	return pwd
}
