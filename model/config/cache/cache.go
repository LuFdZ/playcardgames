package cache

import (
	"encoding/json"
	"fmt"
	mdc "playcards/model/config/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strings"

	"gopkg.in/redis.v5"
)

func ConfigHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("CONFIGS"))
}

func ConfigHSubKey(itemid int32, channel string, version string, mobileos string) string {
	return fmt.Sprintf(cache.KeyPrefix("CONFIGS:%d:%s:%s:%s"), itemid, channel, version, mobileos)
}

func UserHKeySearchList(hSubKey string) map[string]string {
	conditionMap := make(map[string]string)
	hSubKeys := strings.Split(hSubKey, ":")
	if len(hSubKeys[3]) > 0 {
		conditionMap["channel"] = hSubKeys[3]
	}
	if len(hSubKeys[4]) > 0 {
		conditionMap["version"] = hSubKeys[4]
	}
	if len(hSubKeys[5]) > 0 {
		conditionMap["mobileos"] = hSubKeys[5]
	}
	return conditionMap
}

func SetConfigs(cos []*mdc.Config) error {
	err := DeleteConfig()
	if err != nil {
		return err
	}
	var key string
	f := func(tx *redis.Tx) error {
		key = ConfigHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, co := range cos {
				subkey := ConfigHSubKey(co.ItemID, co.Channel, co.Version, co.MobileOs)
				co.Hkey = subkey
				c, _ := json.Marshal(co)
				tx.HSet(key, subkey, string(c))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set config list failed", err)
	}
	log.Info("redis reset all configs")
	return nil
}

func SetConfig(co *mdc.Config) error {
	err := DeleteConfig()
	if err != nil {
		return err
	}
	var key string
	f := func(tx *redis.Tx) error {
		key = ConfigHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			c, _ := json.Marshal(co)
			subkey := ConfigHSubKey(co.ItemID, co.Channel, co.Version, co.MobileOs)
			tx.HSet(key, subkey, string(c))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set config failed", err)
	}
	return nil
}

func GetConfig(itemid int32, channel string, version string, mobileos string) (*mdc.Config, error) {
	key := ConfigHKey()
	subkey := ConfigHSubKey(itemid, channel, version, mobileos)
	//fmt.Printf("AAAGetConfig:%s|%s\n", key, subkey)
	val, err := cache.KV().HGet(key, subkey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get config failed", err)
	}

	co := &mdc.Config{}
	if err := json.Unmarshal(val, co); err != nil {
		return nil, errors.Internal("get config failed", err)
	}
	//fmt.Printf("BBBGetConfig:%v\n", co)
	return co, nil
}

func GetConfigByKey(subkey string) (*mdc.Config, error) {
	key := ConfigHKey()
	val, err := cache.KV().HGet(key, subkey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get config failed", err)
	}

	co := &mdc.Config{}
	if err := json.Unmarshal(val, co); err != nil {
		return nil, errors.Internal("get config failed", err)
	}
	return co, nil
}

func DeleteConfig() error {
	key := ConfigHKey()
	//fmt.Printf("DeleteConfigKey:%s\n", key)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete config error", err)
	}
	return nil
}

func GetAllConfigKey() ([]string, error) {
	var curson uint64
	var rks []string
	var count int64
	count = 999
	for {
		scan := cache.KV().HScan(ConfigHKey(), curson, "*", count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list config list failed", err)
		}

		curson = cur
		rks = append(rks, keys...)

		if curson == 0 {
			break
		}
	}
	return rks, nil
}

func GetAllConfig(f func(*mdc.Config) bool) map[int32]*mdc.Config {
	cm := make(map[int32]*mdc.Config)
	keys, err := GetAllConfigKey()
	if err != nil {
		log.Err("redis get all room err: %v", err)
	}

	for _, k := range keys {
		co, err := GetConfigByKey(k)
		if err != nil {
			log.Err("redis get room err: %v", err)
		}
		if co == nil {
			continue
		}
		if f != nil && !f(co) {
			continue
		}

		if _, ok := cm[co.ItemID]; ok {
			conditionNowMap := UserHKeySearchList(cm[co.ItemID].Hkey)
			conditionCompareMap := UserHKeySearchList(co.Hkey)
			NowMapLen := len(conditionNowMap)
			CompareMapLen := len(conditionCompareMap)
			if _, ok := conditionCompareMap["channel"]; ok && (len(cm[co.ItemID].Channel) == 0 || CompareMapLen >= NowMapLen) {
				cm[co.ItemID] = co
			} else if _, ok := conditionCompareMap["version"]; ok && len(cm[co.ItemID].Channel) == 0 && (len(cm[co.ItemID].Version) == 0 ||
				CompareMapLen >= NowMapLen) {
				cm[co.ItemID] = co
			} else if _, ok := conditionCompareMap["mobileos"]; ok && len(cm[co.ItemID].Channel) == 0 && (len(cm[co.ItemID].Version) == 0 &&
				len(cm[co.ItemID].MobileOs) == 0 || CompareMapLen >= NowMapLen) {
				cm[co.ItemID] = co
			}
		} else {
			cm[co.ItemID] = co
		}
	}
	return cm
}
