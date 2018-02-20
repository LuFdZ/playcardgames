package cache

import (
	"fmt"
	"playcards/utils/cache"
	mdroom "playcards/model/room/mod"
	cacheroom "playcards/model/room/cache"
	//errr "playcards/model/room/errors"
	enumgroom "playcards/model/goldroom/enum"
	//errgr "playcards/model/goldroom/errors"
	"gopkg.in/redis.v5"
	"encoding/json"
	"playcards/utils/tools"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strings"
)

func RoomKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMMAP"))
}

//func GoldRoomKey() string {
//	return fmt.Sprintf(cache.KeyPrefix("GOLDROOMDMAP"))
//}

func GoldRoomSearchKey() string {
	return fmt.Sprintf(cache.KeyPrefix("GOLDROOMSEARCH"))
}

func GoldRoomSearchHKey(gtype int32, level int32, pwd string) string {
	return fmt.Sprintf("gtype:%d-level:%d-password:%s-", gtype, level, pwd)
}

func UserKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMUSER"))
}

func RoomLockKey(rid int32) string {
	return fmt.Sprintf("ROOMLOCK:%d", rid)
}

func SetRoom(mdr *mdroom.Room) error {
	lockKey := RoomLockKey(mdr.RoomID)
	roomKey := RoomKey()
	searchKey := GoldRoomSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(mdr)
			tx.HSet(roomKey, mdr.Password, string(b))
			//g, _ := json.Marshal(mdgroom)
			tx.HSet(searchKey, GoldRoomSearchHKey(mdr.GameType, mdr.Level, mdr.Password),
				fmt.Sprintf("%s-%d-%d", mdr.Password, mdr.Status, enumgroom.NoFull))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("set gold redis gold room failed", err)
	}
	return nil
}

//func GetRoom(rid int32) (*mdroom.Room, error) {
//	key := GoldRoomKey()
//	val, err := cache.KV().HGet(key, tools.IntToString(rid)).Bytes()
//	if err == redis.Nil {
//		return nil, errr.ErrRoomNotExisted //errors.Internal("room not find", err)
//	}
//
//	if err != nil && err != redis.Nil {
//		return nil, errors.Internal("get room failed", err)
//	}
//
//	room := &mdroom.Room{}
//	if err := json.Unmarshal(val, room); err != nil {
//		return nil, errors.Internal("get room failed", err)
//	}
//	return room, nil
//}

func UpdateRoom(mdr *mdroom.Room) error {
	lockKey := RoomLockKey(mdr.RoomID)
	roomKey := RoomKey()
	searchKey := GoldRoomSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(mdr)
			tx.HSet(roomKey, mdr.Password, string(b))
			tx.HSet(searchKey, GoldRoomSearchHKey(mdr.GameType, mdr.Level, mdr.Password),
				fmt.Sprintf("%s-%d-%d", mdr.Password, mdr.Status, enumgroom.NoFull))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, lockKey); err != nil {
		log.Err("%s set room failed\n", lockKey)
		return errors.Internal("set room failed", err)
	}
	return nil
}

func DeleteRoom(mdr *mdroom.Room) error {
	lockKey := RoomLockKey(mdr.RoomID)
	roomKey := RoomKey()
	searcKey := GoldRoomSearchKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(roomKey, mdr.Password)
			searcHKey := GoldRoomSearchHKey(mdr.GameType, mdr.Level, mdr.Password)
			tx.HDel(searcKey, searcHKey)
			//for _, uid := range mdr.Ids {
			//	tx.HDel(UserKey(), tools.IntToString(uid))
			//}
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("delete gold room error", err)
	}
	return nil
}

func SelectGRoom(gtype int32, level int32) (*mdroom.Room, error) {
	var curson uint64
	var count int64
	count = 999
	key := GoldRoomSearchKey()
	match := fmt.Sprintf("gtype:%d-level:%d-*", gtype, level)
	for {
		scan := cache.KV().HScan(key, curson, match, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list gold room list failed", err)
		}
		for i, roomStatus := range keysValues {
			if i%2 == 1 {
				str := strings.Split(roomStatus, "-")
				if str[2] == "1" {
					mdr, err := cacheroom.GetRoom(str[0])
					if err != nil {
						log.Err("select groom err str:%s,err:%v", roomStatus, err)
						continue
					}
					return mdr, nil
				}
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return nil, nil
}

//func DeleteRoomUser(uid int32) error {
//	key := UserKey()
//	f := func(tx *redis.Tx) error {
//		tx.Pipelined(func(p *redis.Pipeline) error {
//			tx.HDel(key, tools.IntToString(uid))
//			log.Info("delete room user:%d\n", uid)
//			return nil
//		})
//		return nil
//	}
//	err := cache.KV().Watch(f, key)
//	if err != nil {
//		return errors.Internal("del room user error", err)
//	}
//	return nil
//}

func GetAllGRoom(gtype int32, statusList []int32) []*mdroom.Room {
	var curson uint64
	var rs []*mdroom.Room
	var count int64
	count = 999
	key := GoldRoomSearchKey()
	match := "*"
	if gtype > 0 {
		match = fmt.Sprintf("gtype:%d-*", gtype)
	}
	for {
		scan := cache.KV().HScan(key, curson, match, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			log.Err("list gold room list failed", err)
			continue
		}
		for i, roomStatus := range keysValues {
			if i%2 == 1 {
				str := strings.Split(roomStatus, "-")
				for _, status := range statusList {
					if tools.StringParseInt(str[1]) == status {
						mdr, err := cacheroom.GetRoom(str[0])
						if err != nil {
							log.Err("get gold room by id failed", err)
							break
						}
						rs = append(rs, mdr)
						break
					}
				}
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return rs
}
