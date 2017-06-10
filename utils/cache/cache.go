package cache

import (
	"strconv"

	"gopkg.in/redis.v5"
)

var cache *redis.Client

func Init(host string) {
	cache = NewCache(host)
}

func NewCache(host string) *redis.Client {
	kv := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
	})
	return kv
}

func KV() *redis.Client {
	return cache
}

func KeyPrefix(key string) string {
	return CacheKeyPrefix + key
}

func SliceStringToInt32(l []string) []int32 {
	ids := []int32{}
	for _, i := range l {
		id, err := strconv.Atoi(i)
		if err != nil {
			continue
		}
		ids = append(ids, int32(id))
	}

	return ids
}
