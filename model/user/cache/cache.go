package cache

import (
	"encoding/json"
	"fmt"
	"math"
	enumu "playcards/model/user/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"
	"sort"
	"strconv"
	"strings"

	"github.com/twinj/uuid"
	"gopkg.in/redis.v5"
	"time"
)

func UserHKey(uid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("USER:%d"), uid)
}

func UserWXKey() string {
	return fmt.Sprintf(cache.KeyPrefix("USERWXTOKEN"))
}

//func UserWXHKey(openid string) string {
//	return fmt.Sprintf(cache.KeyPrefix("USERWXTOKEN:%s"), openid)
//}

func UserOnlineHKey() string {
	return cache.KeyPrefix("USERONLINE")
}

func UserToken(uid int32) string {
	return fmt.Sprintf("%d:%s", uid, uuid.NewV4().String())
}

func UserNumberHKey() string {
	return cache.KeyPrefix("USERNUMBER")
}

func UserHKeySearch() string {
	return cache.KeyPrefix("USER:*")
}

func UserIDFromToken(token string) (int32, error) {
	cols := strings.Split(token, ":")
	if len(cols) != 2 {
		return 0, erru.ErrInvalidToken
	}

	n, err := strconv.Atoi(cols[0])
	if err != nil {
		return 0, erru.ErrInvalidToken
	}

	return int32(n), nil
}

//func GetAccessToken(openid string) (string, error) {
//	key := UserHWXKey(openid)
//	val := cache.KV().HGet(key, "accesstoken").Val()
//	return val, nil
//}

func GetRefreshToken(openid string) (string, error) {
	//key := UserHWXKey(openid)
	key := UserWXKey()
	val := cache.KV().HGet(key, openid).Val()
	return val, nil
}

func SetUserWXInfo(openid string, refreshtoken string) error {
	key := UserWXKey()
	//hKey := UserHWXKey(openid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, openid, refreshtoken)
			//tx.HSet(key, "accesstoken", accesstoken)
			//tx.HSet(key, "refreshtoken", refreshtoken)
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user wx info error", err)
	}

	return nil
}

func GetUser(token string) (*mdu.User, error) {
	uid, err := UserIDFromToken(token)
	if err != nil {
		return nil, err
	}

	key := UserHKey(uid)
	val, err := cache.KV().HGet(key, token).Bytes()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get user failed", err)
	}

	user := &mdu.User{}
	if err := json.Unmarshal(val, user); err != nil {
		return nil, errors.Internal("get user failed", err)
	}
	//user.EncodNickName()
	return user, nil
}

func SetUser(u *mdu.User) (string, error) {
	token := UserToken(u.UserID)
	return token, UpdateUser(token, u)
}

func UpdateUser(token string, u *mdu.User) error {
	key := UserHKey(u.UserID)
	//u.EncodNickName()

	f := func(tx *redis.Tx) error {
		orig, _ := tx.HGet(key, "token").Bytes()
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, string(orig))
			tx.HSet(key, "token", token)
			b, _ := json.Marshal(u)
			tx.HSet(key, token, string(b))
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user error", err)
	}
	return nil
}

func UpdateUserHeartbeat(uid int32) error {
	key := UserHKey(uid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "heartbeat", time.Now().Unix())
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user heartbeat error", err)
	}
	return nil
}

func GetUserHeartbeat(uid int32) (int64, error) {
	key := UserHKey(uid)
	val, err := cache.KV().HGet(key, "heartbeat").Int64()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func SimpleUpdateUser(u *mdu.User) error {
	key := UserHKey(u.UserID)
	val := cache.KV().HGetAll(key).Val()
	token, ok := val["token"]
	if !ok {
		return nil
	}
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(u)
			tx.HSet(key, token, string(b))
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set simple user error", err)
	}
	return nil
}

func GetUserByID(uid int32) (string, *mdu.User) {
	key := UserHKey(uid)
	val := cache.KV().HGetAll(key).Val()
	token, ok := val["token"]
	if !ok {
		return "", nil
	}

	b, ok := val[token]
	if !ok {
		return "", nil
	}

	user := &mdu.User{}
	if err := json.Unmarshal([]byte(b), user); err != nil {
		return "", nil
	}

	return token, user
}

func ListUserByID(uids []int32) []*mdu.User {
	us := make([]*mdu.User, 0, len(uids))

	for _, uid := range uids {
		_, u := GetUserByID(uid)
		if u == nil {
			continue
		}

		us = append(us, u)
	}

	return us
}

func ListUserHKeys() ([]string, error) {
	var curson uint64
	var uks []string
	var count int64
	count = 100

	for {
		scan := cache.KV().Scan(curson, UserHKeySearch(), count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list users failed", err)
		}

		curson = cur
		uks = append(uks, keys...)

		if curson == 0 {
			break
		}
	}
	sort.Strings(uks)
	//fmt.Printf("ListUserHKeys:%v\n",uks)
	return uks, nil
}

func CountUserHKeys() (int32, error) {
	rs, err := ListUserHKeys()
	if err != nil {
		return 0, err
	}
	return int32(len(rs)), err
}

func ListUsers() ([]*mdu.User, error) {
	var us []*mdu.User

	rs, err := ListUserHKeys()
	if err != nil {
		return nil, err
	}

	for _, k := range rs {
		cols := strings.Split(k, ":")
		if len(cols) != 2 {
			return nil, erru.ErrInvalidToken
		}

		uid, err := strconv.Atoi(cols[1])
		if err != nil {
			return nil, errors.Internal("list online users failed", err)
		}

		_, u := GetUserByID(int32(uid))
		if u != nil {
			us = append(us, u)
		}
	}
	return us, nil
}

func GetUserHeartbeats() map[int32]int64 {
	m := make(map[int32]int64)
	keys, err := ListUserHKeys()
	if err != nil {
		log.Err("redis get all user err: %v", err)
	}
	for _, k := range keys {
		cols := strings.Split(k, ":")
		if len(cols) != 3 {
			log.Err("redis get all user token err: %+v", cols)
			continue
		}
		uid, err := strconv.Atoi(cols[2])
		if err != nil {
			log.Err("list strconv users str failed,err:%v", err)
			continue
		}
		userID := int32(uid)
		if GetUserOnlineStatus(userID) == enumu.UserUnline {
			continue
		}
		hearbeat, err := GetUserHeartbeat(userID)
		if err != nil {
			if err == redis.Nil{
				UpdateUserHeartbeat(userID)
			}
			log.Err("list user heartbeat failed,uid:%d,err:%v", userID, err)
			continue
		}
		m[userID] = hearbeat
	}
	return m
}

func GetUserList(f func(*mdu.User) bool, page int32) ([]*mdu.User, int32) {
	var us []*mdu.User
	keys, err := ListUserHKeys()
	if err != nil {
		log.Err("redis get all user err: %v", err)
	}
	//fmt.Printf("GetUserList:%v\n",keys)

	for _, k := range keys {
		//index := i+1
		//if index <= pageStart || index > pageEnd {
		//	continue
		//}
		cols := strings.Split(k, ":")
		if len(cols) != 3 {
			log.Err("redis get all user token err: %+v", cols)
			continue
		}
		uid, err := strconv.Atoi(cols[2])
		if err != nil {
			log.Err("list online users failed,err:%v", err)
			continue
		}
		_, u := GetUserByID(int32(uid))
		if u == nil {
			continue
		}
		if f != nil && !f(u) {
			continue
		}
		us = append(us, u)
	}

	var pageList []*mdu.User
	count := float64(len(us)) / float64(enumu.MaxUserRecordCount)
	count = math.Ceil(count)
	if count == 0 {
		count = 1
	}
	pageStart := 0
	pageStart = int((page - 1) * enumu.MaxUserRecordCount)
	pageEnd := int(pageStart + enumu.MaxUserRecordCount)
	if pageStart > int(count) {
		return nil, int32(count)
	}

	for i, u := range us {
		index := i + 1
		if index <= pageStart || index > pageEnd {
			continue
		}
		pageList = append(pageList, u)
	}
	return pageList, int32(count)
}

func SetUserOnlineStatus(uid int32, status int) error {
	key := UserOnlineHKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			index := uid - 100000 - 1
			tx.SetBit(key, int64(index), status)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user wx info error", err)
	}
	return nil
}

func GetUserOnlineStatus(uid int32) int32 {
	key := UserOnlineHKey()
	index := uid - 100000 - 1
	status := cache.KV().GetBit(key, int64(index))
	online := enumu.UserOnline
	if int32(status.Val()) == 0 {
		online = enumu.UserUnline
	}
	//fmt.Printf("GetUserOnlineStatus:%d|%d\n",uid,online)
	return int32(online)
}

func GetAllOnlineCount() int32 {
	key := UserOnlineHKey()
	//userNumber, err := CountUserHKeys()
	//if err != nil {
	//	return 0, err
	//}
	userNumber := GetUserNumber()
	bc := redis.BitCount{0, userNumber}
	count := cache.KV().BitCount(key, &bc).Val()
	return int32(count)
}

func SetUserNumber(count int32) error {
	key := UserNumberHKey()

	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "count", count)
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user number info error", err)
	}
	return nil
}

func GetUserNumber() int64 {
	key := UserNumberHKey()
	val := cache.KV().HGet(key, "count").Val()
	count, _ := strconv.ParseInt(val, 10, 32)
	return count
}
