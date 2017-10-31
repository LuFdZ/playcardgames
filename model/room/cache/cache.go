package cache

import (
	"encoding/json"
	"fmt"
	mdr "playcards/model/room/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strconv"
	//"strconv"
	"runtime/debug"
	"strings"

	"gopkg.in/redis.v5"
)

func RoomHKey(pwd string) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOM:%s"), pwd)
}

func UserHKey(uid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMUSER:%d"), uid)
}

func RoomHKeySearch() string {
	return cache.KeyPrefix("ROOM:*")
}

func RoomHKeyDelete(gametype int32, rid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMDELETE:%d:%s"), gametype, rid)
}

func RoomHKeyDeleteSearch(gametype int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMDELETE:%d"), gametype)
}

func SetRoom(r *mdr.Room) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = RoomHKey(r.Password)
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(r)
			tx.HSet(key, "password", r.Password)
			tx.HSet(key, "roomid", r.RoomID)
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
		//orig, _ := tx.HGet(key, "password").Bytes()
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete room error", err)
	}
	return nil
}

func CheckRoomExist(pwd string) (bool, error) {
	key := RoomHKey(pwd)
	_, err := cache.KV().HGet(key, "room").Bytes()
	if err == redis.Nil {
		return false, nil
	}
	return true, nil
}

func GetRoom(pwd string) (*mdr.Room, error) {
	key := RoomHKey(pwd)
	val, err := cache.KV().HGet(key, "room").Bytes()
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

func GetRoomByKey(key string) (*mdr.Room, error) {
	val, err := cache.KV().HGet(key, "room").Bytes()
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
			tx.HSet(key, "socketstatus", 1) //scoket连接状态 1在线 2掉线
			tx.HSet(key, "socketnotice", 0) //连接信息是否已广播
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set room user failed", err)
	}
	return nil
}

func UpdateRoomUserSocektStatus(uid int32, socketStatus int32, socketNotice int32) error {
	key := UserHKey(uid)
	rid := cache.KV().HGet(key, "roomid").Val()
	if len(rid) == 0 {
		return nil
	}

	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "socketstatus", socketStatus) //scoket连接状态 1在线 2掉线
			tx.HSet(key, "socketnotice", socketNotice) //scoket连接状态 1在线 2掉线
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set room user failed", err)
	}

	return nil
}

func DeleteRoomUser(uid int32) error {
	key := UserHKey(uid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			log.Info("delete room user:%s\n", key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del room user error", err)
	}
	return nil
}

func DeleteAllRoomUser(password string, callFrom string) error {
	key := RoomHKey(password)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			room, err := GetRoom(password)
			//fmt.Printf("delete room user room:%v \n", room)
			str := ""
			if err == nil && room != nil {
				for _, user := range room.Users {
					userkey := UserHKey(user.UserID)
					rid := tx.HGet(userkey, "roomid").Val()
					roomid, _ := strconv.Atoi(rid)
					if int32(roomid) == room.RoomID {
						tx.Del(userkey)
						str += fmt.Sprintf("|delthisroomuser:%s,roomid:%d|", userkey, roomid)
					} else {
						str += fmt.Sprintf("|nothisroomuser:%s,roomid:%d|", userkey, roomid)
					}
				}
			}
			log.Info("%s DeleteRoomAllUser RoomID:%s,RoomPWd:%s,userlist:%s,Stack:\n%s\n", callFrom, room.RoomID, key, room.RoundNow, str, string(debug.Stack()))
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

func GetRoomPasswordByUserID(uid int32) string {
	key := UserHKey(uid)
	pwd := cache.KV().HGet(key, "password").Val()
	return pwd
}

func GetUserStatus(uid int32) int32 {
	key := UserHKey(uid)
	status := cache.KV().HGet(key, "socketstatus").Val()

	if len(status) > 0 {
		result, _ := strconv.Atoi(status)
		return int32(result)
	}
	return 0
}

func GetUserSocketNotice(uid int32) int32 {
	key := UserHKey(uid)
	status := cache.KV().HGet(key, "socketnotice").Val()
	if len(status) > 0 {
		result, _ := strconv.Atoi(status)
		return int32(result)
	}
	return 0
}

func SetRoomDelete(gametype int32, rid int32) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = RoomHKeyDelete(gametype, rid)
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "roomid", rid)
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set delete room password failed", err)
	}
	return nil
}

func CleanDeleteRoom(gametype int32, rid int32) error {
	key := RoomHKeyDelete(gametype, rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			log.Info("clean delete room user:%s\n", key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del room user error", err)
	}
	return nil
}

func GetAllDeleteRoomKey(gametype int32) ([]int32, error) {
	var curson uint64
	var drks []int32
	var count int64
	count = 999
	for {
		scan := cache.KV().Scan(curson, RoomHKeyDeleteSearch(gametype), count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list delete room id list failed", err)
		}
		curson = cur
		var rids []int32
		for _, key := range keys {
			cols := strings.Split(key, ":")
			if len(cols) != 3 {
				log.Err("get all delete room id format err:%s", key)
				continue
			}
			rid, err := strconv.Atoi(cols[2])
			if err != nil {
				log.Err("get all delete room id no number err:%s", key)
				continue
			}
			rids = append(rids, int32(rid))
		}
		drks = append(drks, rids...)
		if curson == 0 {
			break
		}
	}
	return drks, nil
}

func FlushAll() {
	cache.KV().FlushAll()
}

func GetAllRoomKey() ([]string, error) {
	var curson uint64
	var rks []string
	var count int64
	count = 999
	for {
		scan := cache.KV().Scan(curson, RoomHKeySearch(), count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list room list failed", err)
		}

		curson = cur
		rks = append(rks, keys...)

		if curson == 0 {
			break
		}
	}
	return rks, nil
}

func GetAllRoom(f func(*mdr.Room) bool) []*mdr.Room {
	var rooms []*mdr.Room
	keys, err := GetAllRoomKey()
	if err != nil {
		log.Err("redis get all room err: %v", err)
	}
	for _, k := range keys {
		room, err := GetRoomByKey(k)
		if err != nil {
			log.Err("redis get room err: %v", err)
		}
		if room == nil {
			continue
		}
		if f != nil && !f(room) {
			continue
		}
		rooms = append(rooms, room)
	}
	return rooms
}
