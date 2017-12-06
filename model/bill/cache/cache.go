package cache

import (
	"encoding/json"
	"fmt"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"gopkg.in/redis.v5"
)

func UserBalanceHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("USERBALANCE"))
}

func UserSubBalanceHKey(uid int32,cointype int32) string {
	return fmt.Sprintf(cache.KeyPrefix("%d:%d"),uid,cointype)
}

func GetUserBalance(uid int32,cointype int32) (*mdbill.Balance, error) {
	key := UserBalanceHKey()
	subKey := UserSubBalanceHKey(uid,cointype)
	val, err := cache.KV().HGet(key, subKey).Bytes()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get user balance failed", err)
	}

	balabce := &mdbill.Balance{}
	if err := json.Unmarshal(val, balabce); err != nil {
		return nil, errors.Internal("get user balance failed", err)
	}
	return balabce, nil
}

func SetUserBalance(uid int32,balance *mdbill.Balance) error {
	key := UserBalanceHKey()
	subKey := UserSubBalanceHKey(uid,balance.CoinType)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(balance)
			tx.HSet(key, subKey, string(b))
			return nil
		})
		return nil
	}

	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set user balance error", err)
	}
	return nil
}
