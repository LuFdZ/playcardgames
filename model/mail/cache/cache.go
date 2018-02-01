package cache

import (
	"encoding/json"
	"fmt"
	mdgame "playcards/model/mail/mod"
	enumgame "playcards/model/mail/enum"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	"gopkg.in/redis.v5"
	"playcards/utils/tools"
	"strings"
	"math"
)

func MailInfoHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("MAILINFOMAP"))
}

func MailSendLogHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("MAILSENDLOGMAP"))
}

func PlayerMailHKey() string {
	return fmt.Sprintf(cache.KeyPrefix("PLAYERMAILMAP"))
}

func PlayerMailHSubKey(uid int32, logid int32) string {
	return fmt.Sprintf("uid:%d-logid:%d-", uid, logid)
}

func SetMailInfos(mis []*mdgame.MailInfo) error {
	key := MailInfoHKey()
	err := DeleteAllMails(key)
	if err != nil {
		return err
	}

	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, mi := range mis {
				b, _ := json.Marshal(mi)
				tx.HSet(key, tools.IntToString(mi.MailID), string(b))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set mail info list failed", err)
	}
	log.Info("redis reset all mail info")
	return nil
}

func SetMailSendLogs(mlss []*mdgame.MailSendLog) error {
	key := MailSendLogHKey()
	err := DeleteAllMails(key)
	if err != nil {
		return err
	}
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, mls := range mlss {
				b, _ := json.Marshal(mls)
				tx.HSet(key, tools.IntToString(mls.LogID), string(b))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set mail send log list failed", err)
	}
	log.Info("redis reset all mail send log info")
	return nil
}

func SetPlayerMails(pms []*mdgame.PlayerMail) error {
	key := PlayerMailHKey()
	err := DeleteAllMails(key)
	if err != nil {
		return err
	}
	f := func(tx *redis.Tx) error {
		key = PlayerMailHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, pm := range pms {
				b, _ := json.Marshal(pm)
				tx.HSet(key, tools.IntToString(pm.LogID), string(b))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set player mail list failed", err)
	}
	log.Info("redis reset all player mail info")
	return nil
}

func GetMailInfo(mid int32) (*mdgame.MailInfo, error) {
	key := MailInfoHKey()
	val, err := cache.KV().HGet(key, tools.IntToString(mid)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get mail info failed", err)
	}

	mi := &mdgame.MailInfo{}
	if err := json.Unmarshal(val, mi); err != nil {
		return nil, errors.Internal("get mail info failed", err)
	}
	//fmt.Printf("BBBGetConfig:%v\n", co)
	return mi, nil
}

func DeleteAllMails(key string) error {
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete mail info error", err)
	}
	return nil
}

func SetMailSendLog(msl *mdgame.MailSendLog) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = MailSendLogHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(msl)
			tx.HSet(key, tools.IntToString(msl.LogID), string(b))
			//expire := enumgame.MailEndLogOverTime
			//tx.Expire(key, expire)
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set mail send log list failed id:%d,err:", err)
	}
	return nil
}

func GetMailSendLog(logID int32) (*mdgame.MailSendLog, error) {
	key := MailSendLogHKey()
	val, err := cache.KV().HGet(key, tools.IntToString(logID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get mail send log failed", err)
	}

	co := &mdgame.MailSendLog{}
	if err := json.Unmarshal(val, co); err != nil {
		return nil, errors.Internal("get mail send log failed", err)
	}
	return co, nil
}

func DeleteMailSendLog(logID int32) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = MailSendLogHKey()
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, tools.IntToString(logID))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set mail send log list failed id:%d,err:", err)
	}
	return nil
}

func SetPlayerMail(pm *mdgame.PlayerMail) error {
	key := PlayerMailHKey()
	f := func(tx *redis.Tx) error {
		subkey := PlayerMailHSubKey(pm.UserID, pm.LogID)
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(pm)
			tx.HSet(key, subkey, string(b))
			expire := enumgame.PlayerMailOverTime
			tx.Expire(key, expire)
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set player mail list failed err:", err)
	}
	return nil
}

func UpdatePlayerMail(pm *mdgame.PlayerMail) error {
	key := PlayerMailHKey()
	f := func(tx *redis.Tx) error {
		subkey := PlayerMailHSubKey(pm.UserID, pm.LogID)
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(pm)
			tx.HSet(key, subkey, string(b))
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("update player mail list failed err:", err)
	}
	return nil
}

func GetPlayerMail(uid int32, logID int32) (*mdgame.PlayerMail, error) {
	key := PlayerMailHKey()
	subkey := PlayerMailHSubKey(uid, logID)
	val, err := cache.KV().HGet(key, subkey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get player mail failed", err)
	}

	pm := &mdgame.PlayerMail{}
	if err := json.Unmarshal(val, pm); err != nil {
		return nil, errors.Internal("get player mail failed", err)
	}
	return pm, nil
}

func DeletePlayerMail(uid int32, logID int32) error {
	var key string
	f := func(tx *redis.Tx) error {
		key = PlayerMailHKey()
		subkey := PlayerMailHSubKey(uid, logID)
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, subkey)
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("delete player mail failed,err:", err)
	}
	return nil
}

func GetPlayerMails(uid int32) ([]*mdgame.PlayerMail, error) {
	var curson uint64
	var pms []*mdgame.PlayerMail
	var count int64
	count = 999
	key := PlayerMailHKey()
	for {
		search := fmt.Sprintf("%d:*", uid)
		scan := cache.KV().HScan(key, curson, search, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list player mail list failed", err)
		}
		var logID int32
		for i, idStr := range keysValues {
			if i%2 == 1 {
				logID = tools.StringParseInt(strings.Split(idStr, ":")[1])
			} else if i%2 == 0 {
				pm, err := GetPlayerMail(uid, logID)
				if err != nil {
					log.Err("get player mails rid:%s,err:%v", idStr, err)
					continue
				}
				pms = append(pms, pm)
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}
	return pms, nil
}

func PagePlayerMailList(page int32, pms []*mdgame.PlayerMail) ([]*mdgame.PlayerMail, int32, int32) {
	total := int32(len(pms))
	var pageList []*mdgame.PlayerMail
	count := float64(len(pms)) / float64(enumgame.MaxMailRecordCount)
	count = math.Ceil(count)
	if count == 0 {
		count = 1
	}
	pageStart := 0
	pageStart = int((page - 1) * enumgame.MaxMailRecordCount)
	pageEnd := int(pageStart + enumgame.MaxMailRecordCount)
	if pageStart > int(count) {
		return nil, int32(count), total
	}
	for i, pm := range pms {
		index := i + 1
		if index <= pageStart || index > pageEnd {
			continue
		}
		pageList = append(pageList, pm)
	}

	return pageList, int32(count), total
}

func GetAndRefreshPlayerMailByID(uid int32) []*mdgame.PlayerMail {
	var curson uint64
	var pms []*mdgame.PlayerMail
	var count int64
	count = 999
	key := PlayerMailHKey()
	match := fmt.Sprintf("uid:%d-*", uid)

	for {
		scan := cache.KV().HScan(key, curson, match, count)
		keysValues, cur, err := scan.Result()
		if err != nil {
			log.Err("list player mail list failed", err)
			continue
		}
		for i, pmStr := range keysValues {
			if i%2 == 1 {
				pm := &mdgame.PlayerMail{}
				if err := json.Unmarshal([]byte(pmStr), &pm); err != nil {
					log.Err("get player mail unmarshal err str:%s,err:%v", pmStr, err)
					continue
				}
				pms = append(pms, pm)
			}
		}
		curson = cur
		if curson == 0 {
			break
		}
	}

	return pms
}

func GetPlayerMailByID(page int32, uid int32) ([]*mdgame.PlayerMail, int32, int32) {
	pms := GetAndRefreshPlayerMailByID(uid)
	return PagePlayerMailList(page, pms)
}
