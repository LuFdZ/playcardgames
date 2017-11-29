package cache

import (
	"encoding/json"
	"fmt"
	mdCommon "playcards/model/common/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"
	"playcards/utils/log"

	redis "gopkg.in/redis.v5"
)

func BlackListHKey(typeid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("BLACKLIST:%d"), typeid)
}

func BlackListSubHKey(originid int32, targetid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("SUB:%d:%d"), originid, targetid)
}

func ExamineHKey(typeid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("EXAMINE:%d"), typeid)
}

func ExamineSubHKey(auditor int32, applicant int32) string {
	return fmt.Sprintf(cache.KeyPrefix("SUB:%d:%d"), auditor, applicant)
}

func SetBlackList(mbl *mdCommon.BlackList) error {
	key := BlackListHKey(mbl.Type)
	subKey := BlackListSubHKey(mbl.OriginID, mbl.TargetID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(mbl)
			tx.HSet(key, subKey, string(b))
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set redis black list failed", err)
	}
	log.Err("set redis black list type:%s,orgonid:%d,targetid:%d\n", mbl.Type, mbl.OriginID, mbl.TargetID)
	return nil
}

func GetBlackList(typeid int32, originid int32, targetid int32) (*mdCommon.BlackList, error) {
	key := BlackListHKey(typeid)
	subKey := BlackListSubHKey(originid, targetid)
	val, err := cache.KV().HGet(key, subKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get redis black list failed", err)
	}
	bl := &mdCommon.BlackList{}
	if err := json.Unmarshal(val, bl); err != nil {
		return nil, errors.Internal("get redis black list failed", err)
	}
	return bl, nil
}

func DeleteBlackList(mbl *mdCommon.BlackList) error {
	key := BlackListHKey(mbl.Type)
	subKey := BlackListSubHKey(mbl.OriginID, mbl.TargetID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, subKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete redis black list failed", err)
	}
	log.Err("delete redis black list type:%s,orgonid:%d,targetid:%d\n", mbl.Type, mbl.OriginID, mbl.TargetID)
	return nil
}

func SetExamine(me *mdCommon.Examine) error {
	key := ExamineHKey(me.Type)
	subKey := ExamineSubHKey(me.AuditorID, me.ApplicantID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(me)
			tx.HSet(key, subKey, string(b))
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set redis examine failed", err)
	}
	log.Err("set redis black list type:%s,auditor:%d,applicant:%d\n", me.Type, me.AuditorID, me.ApplicantID)
	return nil
}

func GetExamine(typeid int32, auditorid int32, applicantid int32) (*mdCommon.Examine, error) {
	key := ExamineHKey(typeid)
	subKey := ExamineSubHKey(auditorid, applicantid)
	val, err := cache.KV().HGet(key, subKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get redis black list failed", err)
	}
	me := &mdCommon.Examine{}
	if err := json.Unmarshal(val, me); err != nil {
		return nil, errors.Internal("get redis black list failed", err)
	}
	return me, nil
}

func DeleteExamine(me *mdCommon.Examine) error {
	key := ExamineHKey(me.Type)
	subKey := ExamineSubHKey(me.AuditorID, me.ApplicantID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, subKey)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete redis black list failed", err)
	}
	log.Err("delete redis black list type:%s,auditorid:%d,applicantid:%d\n", me.Type, me.AuditorID, me.ApplicantID)
	return nil
}

func DeleteAll(key string) error {
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.Del(key)
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal(fmt.Sprintf("delete all commonSrv error,key:%s", key), err)
	}
	return nil
}

func SetAllBlackList(typeid int32, mbls []*mdCommon.BlackList) error {
	key := BlackListHKey(typeid)
	err := DeleteAll(key)
	if err != nil {
		return err
	}
	f := func(tx *redis.Tx) error {

		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, mbl := range mbls {
				subKey := BlackListSubHKey(mbl.OriginID, mbl.TargetID)
				c, _ := json.Marshal(mbl)
				tx.HSet(key, subKey, string(c))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set black lists failed", err)
	}
	log.Err("redis reset all black list")
	return nil
}

func SetAllExamine(typeid int32, mes []*mdCommon.Examine) error {
	key := ExamineHKey(typeid)
	err := DeleteAll(key)
	if err != nil {
		return err
	}
	f := func(tx *redis.Tx) error {

		tx.Pipelined(func(p *redis.Pipeline) error {
			for _, me := range mes {
				subKey := ExamineSubHKey(me.AuditorID, me.ApplicantID)
				c, _ := json.Marshal(me)
				tx.HSet(key, subKey, string(c))
			}
			return nil
		})
		return nil
	}
	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set examine list failed", err)
	}
	log.Err("redis reset all examines")
	return nil
}
