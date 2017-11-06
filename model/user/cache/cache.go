package cache

import (
	"encoding/json"
	"fmt"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"strconv"
	"strings"

	"github.com/twinj/uuid"
	"gopkg.in/redis.v5"
)

func UserHKey(uid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("USER:%d"), uid)
}

func UserHWXKey(openid string) string {
	return fmt.Sprintf(cache.KeyPrefix("USER:%d"), openid)
}

func UserHKeySearch() string {
	return cache.KeyPrefix("USER:*")
}

func UserToken(uid int32) string {
	return fmt.Sprintf("%d:%s", uid, uuid.NewV4().String())
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

func GetAccessToken(openid string) (string, error) {
	key := UserHWXKey(openid)
	val := cache.KV().HGet(key, "accesstoken").Val()
	return val, nil
}

func GetRefreshToken(openid string) (string, error) {
	key := UserHWXKey(openid)
	val := cache.KV().HGet(key, "refreshtoken").Val()
	return val, nil
}

func SetUserWXInfo(openid string, accesstoken string, refreshtoken string) error {
	key := UserHWXKey(openid)

	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HSet(key, "accesstoken", accesstoken)
			tx.HSet(key, "refreshtoken", refreshtoken)
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
