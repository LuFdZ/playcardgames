package cache

import (
	"encoding/json"
	"fmt"
	"math"
	enumr "playcards/model/room/enum"
	errr "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strconv"
	"strings"

	"gopkg.in/redis.v5"
	"playcards/utils/tools"
)

//var roomMap = make(map[string]mdr.Room)

func AgentRoomHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("AGENTROOM"))
}

func AgentRoomHSubKey(uid int32, gametype int32, rid int32, pwd string) string {
	return fmt.Sprintf("uid:%d-gametype:%d-rid:%d-pwd:%s-", uid, gametype, rid, pwd)
}

func RoomHKeyDelete(gametype int32, rid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMDELETE:%d:%d"), gametype, rid)
}

func RoomHKeyDeleteSearch(gametype int32) string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMDELETE:%d*"), gametype)
}

func RoomLockKey(password string) string {
	return fmt.Sprintf("ROOMLOCK:%s", password)
}

func RoomKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMMAP"))
}

func UserKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMUSER"))
}

func RoomSearcKey() string {
	return fmt.Sprintf(cache.KeyPrefix("ROOMSEARCH"))
}

func RoomSearchHKey(gtype int32, status int32, password string) string {
	return fmt.Sprintf("gtype:%d-status:%d-password:%s-", gtype, status, password)
}

func SetRoom(mdroom *mdr.Room) error {
	lockKey := RoomLockKey(mdroom.Password)
	roomKey := RoomKey()
	searchKey := RoomSearcKey()
	f := func(tx *redis.Tx) error {
		searchHKey := RoomSearchHKey(mdroom.GameType, mdroom.Status, mdroom.Password)
		tx.Pipelined(func(p *redis.Pipeline) error {
			mdroom.SearchKey = searchHKey
			b, _ := json.Marshal(mdroom)
			tx.HSet(roomKey, mdroom.Password, string(b))
			tx.HSet(searchKey, searchHKey, mdroom.Password)
			if mdroom.RoomType == enumr.RoomTypeAgent {
				aKey := AgentRoomHKey()
				subKey := AgentRoomHSubKey(mdroom.PayerID, mdroom.GameType, mdroom.RoomID, mdroom.Password )
				tx.HSet(aKey, subKey,string(b))
			}
			return nil
		})
		//roomMap[mdroom.Password] = *mdroom
		//cache.KV().ZAdd(RoomRankKey(), redis.Z{Score: float64(mdroom.RoomID), Member: mdroom.Password})
		return nil
	}
	if err := cache.KV().Watch(f, lockKey); err != nil {
		return errors.Internal("set room failed", err)
	}
	return nil
}

func UpdateRoom(mdroom *mdr.Room) error {
	lockKey := RoomLockKey(mdroom.Password)
	roomKey := RoomKey()
	searchKey := RoomSearcKey()

	f := func(tx *redis.Tx) error {
		searchHkey := RoomSearchHKey(mdroom.GameType, mdroom.Status, mdroom.Password)
		tx.Pipelined(func(p *redis.Pipeline) error {
			lastKey := mdroom.SearchKey
			tx.HDel(searchKey, lastKey)
			mdroom.SearchKey = searchHkey
			b, _ := json.Marshal(mdroom)
			tx.HSet(roomKey, mdroom.Password, string(b))
			tx.HSet(searchKey, searchHkey, mdroom.Password)
			if mdroom.RoomType == enumr.RoomTypeAgent {
				aKey := AgentRoomHKey()
				subKey := AgentRoomHSubKey(mdroom.PayerID, mdroom.GameType, mdroom.RoomID, mdroom.Password)
				tx.HSet(aKey, subKey, string(b))
			}
			return nil
		})
		//roomMap[mdroom.Password] = *mdroom
		return nil
	}
	if err := cache.KV().Watch(f, lockKey); err != nil {
		log.Err("%s set room failed\n",lockKey)
		return errors.Internal("set room failed", err)
	}

	return nil
}

func DeleteRoom(mdroom *mdr.Room) error {
	lockKey := RoomLockKey(mdroom.Password)
	roomKey := RoomKey()
	searcKey := RoomSearcKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(roomKey, mdroom.Password)
			tx.HDel(searcKey, mdroom.SearchKey)
			//tx.ZRem(rankKey, mdroom.Password)
			return nil
		})
		//delete(roomMap, pwd)
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("delete room error", err)
	}
	return nil
}

//func deleteRoomSearch(tx *redis.Tx,pwd string){
//	key := GetRoomKey(pwd)
//	tx.HDel(RoomSearcKey(),key)
//}

func CheckRoomExist(pwd string) bool {
	key := RoomKey()
	return cache.KV().HExists(key, pwd).Val()
}

func GetRoom(pwd string) (*mdr.Room, error) {
	key := RoomKey()
	val, err := cache.KV().HGet(key, pwd).Bytes()
	if err == redis.Nil {
		return nil, errr.ErrRoomNotFind //errors.Internal("room not find", err)
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

func GetRoomTestConfigKey(key string) string {
	val := cache.KV().HGet("TESTKEY", key).Val()
	if len(val) == 0 {
		return "0"
	}
	return val
}

func SetRoomUser(rid int32, password string, uid int32) error {
	key := UserKey()
	value := fmt.Sprintf("%s:%d", password, rid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, tools.String2int(uid), value)
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
	key := UserKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.String2int(uid))
			log.Info("delete room user:%d\n", uid)
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

func DeleteAllRoomUser(pwd string, callFrom string) error {
	lockKey := RoomLockKey(pwd)
	userkey := UserKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			room, err := GetRoom(pwd)
			//fmt.Printf("delete room user room:%v \n", room)
			str := ""
			if err == nil && room != nil {
				for _, user := range room.Users {
					value := tx.HGet(userkey, tools.String2int(user.UserID)).Val()
					if len(value) == 0{
						continue
					}
					rid := strings.Split(value, ":")[1]
					roomid, _ := strconv.Atoi(rid)
					if int32(roomid) == room.RoomID {
						tx.HDel(userkey,tools.String2int(user.UserID) )
						str += fmt.Sprintf("|delthisroomuser:%s", user.UserID)
					} else {
						str += fmt.Sprintf("|nothisroomuser:%s,roomid:%d|", user.UserID, roomid)
					}
				}
			}
			log.Info("%s DeleteRoomAllUser RoomID:%s,RoomPWd:%s,user list:%s", callFrom, room.RoomID, lockKey, room.RoundNow, str)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set room error", err)
	}
	return nil
}

//func GetRoomUserID(uid int32) (*mdr.Room, error) {
//	key := UserKey()
//	value := cache.KV().HGet(key, string(uid)).Val()
//	if len(value) == 0 {
//		return nil, errr.ErrUserNotInRoom
//	}
//	roomInfo := strings.Split(value, ":")
//	pwd := roomInfo[0]
//	rid, _ := strconv.Atoi(roomInfo[1])
//	mdroom, err := GetRoom(pwd)
//	if err != nil {
//		return nil, err
//	}
//	if mdroom.RoomID != int32(rid) {
//		DeleteRoomUser(uid)
//		return nil, nil
//	}
//	return mdroom, nil
//}

func ExistRoomUser(uid int32) bool {
	key := UserKey()
	return cache.KV().HExists(key, tools.String2int(uid)).Val()
}

func GetRoomUserID(uid int32) (*mdr.Room, error) {
	key := UserKey()
	value := cache.KV().HGet(key, tools.String2int(uid)).Val()
	if len(value) == 0 {
		return nil, errr.ErrUserNotInRoom
	}
	roomInfo := strings.Split(value, ":")
	pwd := roomInfo[0]
	rid, _ := strconv.Atoi(roomInfo[1])
	mdroom, err := GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if mdroom.RoomID != int32(rid) {
		DeleteRoomUser(uid)
		return nil, errr.ErrUserNotInRoom
	}
	return mdroom, nil
}

func SetAgentRoom(r *mdr.Room) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = AgentRoomHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			subKey := AgentRoomHSubKey(r.PayerID, r.GameType, r.RoomID, r.Password)
			b, _ := json.Marshal(r)
			tx.HSet(key, subKey, b)
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set config failed", err)
	}
	return nil
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
			if len(cols) != 4 {
				log.Err("get all delete room id format err:%s", key)
				continue
			}
			rid, err := strconv.Atoi(cols[3])
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
	//fmt.Printf("GetAllDeleteRoomKey:%+v\n",drks)
	return drks, nil
}

func FlushAll() {
	cache.KV().FlushAll()
}

func GetAgentRoom(uid int32, gameType int32, rid int32, pwd string) (*mdr.Room, error) {
	key := AgentRoomHKey()
	subKey := AgentRoomHSubKey(uid, gameType, rid, pwd)
	password := cache.KV().HGet(key, subKey).Val()
	room, err := GetRoom(password)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func DeleteAgentRoom(uid int32, gameType int32, rid int32, pwd string) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = AgentRoomHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			subKey := AgentRoomHSubKey(uid, gameType, rid, pwd)
			//fmt.Printf("DeleteAgentRoom:%s\n",subKey)
			tx.HDel(key, subKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del room user error", err)
	}
	log.Debug("DeleteAgentRoom userid:%d,gametype:%d,rid%d,pwd:%s", uid, gameType, rid, pwd)
	return nil
}

func GetAgentRoomKey(gametype int32, uid int32) ([]string, error) {
	var curson uint64
	var rks []string
	var count int64
	count = 999
	match := ""
	if gametype == enumr.AgentRoomAllGameType { //match = fmt.Sprintf("%d:%d:%d:%s", uid, gametype, rid, pwd)
		match = fmt.Sprintf("uid:%d-*", uid)
	} else {
		match = fmt.Sprintf("uid:%d-gametype:%d-*", uid, gametype)
	}
	for {
		scan := cache.KV().HScan(AgentRoomHKey(), curson, match, count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list room list failed", err)
		}
		//var temp []string
		for i, k := range keys {
			if i%2 == 1 {
				rks = append(rks, k)
			}
			//if k[0] == 123 {
			//	temp = append(temp, k)
			//}
		}
		curson = cur
		//rks = append(rks, temp...)

		if curson == 0 {
			break
		}
	}

	//sort.Strings(rks)
	//fmt.Printf("GetAgentRoomKey:%v\n",rks)
	return rks, nil
}

func PageAgentRoom(uid int32, gametype int32, page int32, f func(*mdr.Room) bool) ([]*mdr.Room, int32, int32) {
	var rooms []*mdr.Room
	keys, err := GetAgentRoomKey(gametype, uid)
	if err != nil {
		log.Err("redis get all room err: %v", err)
	}

	for _, k := range keys {
		room := &mdr.Room{}
		//room, err := GetRoom(k)
		if err = json.Unmarshal([]byte(k), room); err != nil {
			log.Err("redis get room err: str:%s,err:%v", k, err)
			continue
		}
		if room == nil {
			continue
		}
		if f != nil && !f(room) {
			continue
		}
		rooms = append(rooms, room)
	}
	return PageRoomList(page, rooms)
}

//func PageRedisRoom(page int32, f func(*mdr.Room) bool) ([]*mdr.Room, int32, int32) {
//	var rooms []*mdr.Room
//	keys, err := GetAllRoomKey()
//	if err != nil {
//		log.Err("redis get all room err: %v", err)
//	}
//	for _, k := range keys {
//		room, err := GetRoomByKey(k)
//		if err != nil {
//			log.Err("redis get room err: %v", err)
//		}
//		if room == nil {
//			continue
//		}
//		if f != nil && !f(room) {
//			continue
//		}
//		rooms = append(rooms, room)
//	}
//	return PageRoomList(page, rooms)
//
//}

func PageRoomList(page int32, rooms []*mdr.Room) ([]*mdr.Room, int32, int32) {
	total := int32(len(rooms))
	var pageList []*mdr.Room
	count := float64(len(rooms)) / float64(enumr.MaxAgentRoomRecordCount)
	count = math.Ceil(count)
	if count == 0 {
		count = 1
	}
	if page > enumr.AgentRoomAllPage {
		pageStart := 0
		pageStart = int((page - 1) * enumr.MaxAgentRoomRecordCount)
		pageEnd := int(pageStart + enumr.MaxAgentRoomRecordCount)
		if pageStart > int(count) {
			return nil, int32(count), total
		}
		for i, u := range rooms {
			index := i + 1
			if index <= pageStart || index > pageEnd {
				continue
			}
			pageList = append(pageList, u)
		}
	} else {
		pageList = rooms
	}

	return pageList, int32(count), total
}

//func GetAllRoom(f func(int32, int32) bool) []*mdr.Room {
//	var rooms []*mdr.Room
//	//var keys []string
//	for _, rm := range roomMap {
//		if f != nil && !f(rm) {
//			continue
//		}
//		T1 := &rm
//		T2 := *T1
//		mdroom := &T2
//		rooms = append(rooms, mdroom)
//	}
//	return rooms
//}

func GetAllRoomByStatus(status int32) []*mdr.Room {
	var rooms []*mdr.Room
	match := fmt.Sprintf("*status:%d-*", status)
	rooms, err := GetAllRoomKey(match)
	if err != nil {
		log.Err("GetAllRoomByStatus:%v", err)
	}
	return rooms
}

func GetAllRoomByGameTypeAndStatus(gtype int32, status int32) []*mdr.Room {
	var rooms []*mdr.Room
	match := fmt.Sprintf("*gtype:%d-status:%d-*", gtype, status)
	rooms, err := GetAllRoomKey(match)
	if err != nil {
		log.Err("GetAllRoomByStatus:%v", err)
	}
	return rooms
}

func GetAllRooms(f func(*mdr.Room) bool) []*mdr.Room {
	var (
		rooms []*mdr.Room
		out   []*mdr.Room
	)
	match := "*"
	rooms, err := GetAllRoomKey(match)
	if err != nil {
		log.Err("GetAllRoomByStatus:%v", err)
	}
	if len(rooms) == 0 {
		return rooms
	}

	for _, room := range rooms {
		if f != nil && !f(room) {
			continue
		}
		out = append(out, room)
	}
	return out
}

func GetAllRoomKey(match string) ([]*mdr.Room, error) {
	var curson uint64
	var rs []*mdr.Room
	var count int64
	count = 999
	key := RoomSearcKey()
	for {
		scan := cache.KV().HScan(key, curson, match, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list room list failed", err)
		}
		for i, searchRoom := range keysValues {
			if i%2 == 0 {
				password := strings.Split(searchRoom, "-")[2]
				pwd := strings.Split(password, ":")[1]
				room, err := GetRoom(pwd)
				if err != nil {
					log.Err("GetAllRoomKeyErr match:%s,err:%v", match, err)
				}
				rs = append(rs, room)
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return rs, nil
}
