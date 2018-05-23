package cache

import (
	"encoding/json"
	"fmt"
	mdlog "playcards/model/log/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"
	redis "gopkg.in/redis.v5"
	"crypto/md5"
	"encoding/hex"
	"time"
)

func LogKey() string {
	return fmt.Sprintf(cache.KeyPrefix("LOGMAP"))
}

func SetErrLog(scode int32, errStr string) error {
	lockKey := LogKey()
	t := time.Now()
	mdl := &mdlog.ErrLog{}
	h := md5.New()
	h.Write([]byte(errStr))
	subKey := hex.EncodeToString(h.Sum(nil))
	f := func(tx *redis.Tx) error {
		if CheckLogExist(subKey) {
			mdl, _ = GetErrLog(subKey)
			mdl.Times++
		} else {
			mdl.ServerCode = scode
			mdl.Error = errStr
			mdl.Date = &t
			mdl.Times = 0
		}
		tx.Pipelined(func(p *redis.Pipeline) error {
			str, _ := json.Marshal(mdl)
			tx.HSet(lockKey, subKey, str)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, lockKey)
	if err != nil {
		return errors.Internal("set err log failed", err)
	}
	return nil
}

func CheckLogExist(subKey string) bool {
	key := LogKey()
	return cache.KV().HExists(key, subKey).Val()
}

func DeleteGame(subKey string) error {
	key := LogKey()
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, subKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("del err log redis error", err)
	}
	return nil
}

func GetErrLog(subKey string) (*mdlog.ErrLog, error) {
	key := LogKey()
	val, err := cache.KV().HGet(key, subKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get err log failed", err)
	}
	mdl := &mdlog.ErrLog{}
	if err := json.Unmarshal(val, mdl); err != nil {
		return nil, errors.Internal("get err log failed", err)
	}
	return mdl, nil
}

func GetAllErrLog() ([]string, []string) {
	var curson uint64
	var keys []string
	var values []string
	var count int64
	count = 999
	key := LogKey()
	for {
		scan := cache.KV().HScan(key, curson, "*", count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			log.Err("err log list failed", err)
			return nil, nil
		}
		for i, search := range keysValues {
			if i%2 == 0 {
				keys = append(keys, search)
			} else {
				values = append(values, search)
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return keys, values
}
