package cache

import (
	"encoding/json"
	"fmt"
	errclub "playcards/model/club/errors"
	mdclub "playcards/model/club/mod"
	"playcards/utils/cache"
	"playcards/utils/errors"

	redis "gopkg.in/redis.v5"
	//"sort"
	cacheuser "playcards/model/user/cache"
	enumuser "playcards/model/user/enum"
	"playcards/utils/log"
	"strconv"
)

func ClubHKey(cid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("CLUB:%d"), cid)
}

func ClubHKeySearch() string {
	return cache.KeyPrefix("CLUB:*")
}

func ClubMemberHKey(cid int32) string {
	return fmt.Sprintf(cache.KeyPrefix("CLUBMEMBER:%d"), cid)
}
func ClubMemberHKeySearch() string {
	return cache.KeyPrefix("CLUBMEMBER:*")
}

//func ClubMemberSubKey(uid int32) string {
//	return fmt.Sprintf(cache.KeyPrefix("MEMBER:%d"), uid)
//}

func SetClub(mclub *mdclub.Club) error {
	key := ClubHKey(mclub.ClubID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(mclub)
			tx.HSet(key, "club", string(b))
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("set club failed", err)
	}
	return nil
}

func GetClub(clubid int32) (*mdclub.Club, error) {
	key := ClubHKey(clubid)
	val, err := cache.KV().HGet(key, "club").Bytes()
	if err == redis.Nil {
		return nil, errclub.ErrClubNotExisted
	}
	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get club failed", err)
	}
	club := &mdclub.Club{}
	if err := json.Unmarshal(val, club); err != nil {
		return nil, errors.Internal("get club failed", err)
	}
	return club, nil
}

func CheckClubExists(clubid int32) bool {
	key := ClubHKey(clubid)
	return cache.KV().Exists(key).Val()
}

func SetClubMember(mcm *mdclub.ClubMember) error {
	key := ClubMemberHKey(mcm.ClubID)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			b, _ := json.Marshal(mcm)
			tx.HSet(key, strconv.Itoa(int(mcm.UserID)), string(b))
			return nil
		})
		return nil
	}

	if err := cache.KV().Watch(f, key); err != nil {
		return errors.Internal("set club member failed", err)
	}

	return nil
}

func GetClubMember(clubid int32, uid int32) (*mdclub.ClubMember, error) {
	key := ClubMemberHKey(clubid)
	//subKey := ClubMemberSubKey(uid)
	val, err := cache.KV().HGet(key, strconv.Itoa(int(uid))).Bytes()
	if err == redis.Nil {
		return nil, errclub.ErrClubNotExisted
	}

	if err != nil && err != redis.Nil {
		return nil, errors.Internal("get club member failed", err)
	}

	mcm := &mdclub.ClubMember{}
	if err := json.Unmarshal(val, mcm); err != nil {
		return nil, errors.Internal("get club member failed", err)
	}
	return mcm, nil
}

func DeleteClubMember(clubid int32, uid int32) error {
	key := ClubMemberHKey(clubid)
	//subKey := ClubMemberSubKey(uid)
	f := func(tx *redis.Tx) error {
		tx.Pipelined(func(p *redis.Pipeline) error {
			tx.HDel(key, strconv.Itoa(int(uid)))
			return nil
		})
		return nil
	}
	err := cache.KV().Watch(f, key)
	if err != nil {
		return errors.Internal("delete redis club member failed", err)
	}
	log.Err("delete redis club member clubid:%,uid%d\n", clubid, uid)
	return nil
}

func ListClubMemberHKey(clubid int32, online bool) ([]int32, error) {
	var curson uint64
	var cmks []string
	var uks []int32
	var count int64
	count = 100
	for {
		scan := cache.KV().HScan(ClubMemberHKey(clubid), curson, "*", count)
		keys, cur, err := scan.Result()
		if err != nil {
			return nil, errors.Internal("list club member failed", err)
		}

		curson = cur
		cmks = append(cmks, keys...)

		if curson == 0 {
			break
		}
	}
	for _, k := range cmks {
		if k[0] != 123 {
			uk, _ := strconv.Atoi(k)
			uid := int32(uk)
			if online {
				ol := cacheuser.GetUserOnlineStatus(uid)
				if ol == enumuser.UserUnline {
					continue
				}
			}
			uks = append(uks, uid)
		}
	}
	//sort.Ints(uks)
	return uks, nil
}

func CountClubMemberHKeys(clubid int32) (int32, error) {
	cmks, err := ListClubMemberHKey(clubid, false)
	if err != nil {
		return 0, err
	}
	return int32(len(cmks)), err
}

func GetAllClubMember(clubid int32, online bool) []*mdclub.ClubMember {
	var rooms []*mdclub.ClubMember
	keys, err := ListClubMemberHKey(clubid, online)
	if err != nil {
		log.Err("redis get all club member err: %v", err)
	}
	for _, k := range keys {
		uid := int32(k)
		mcm, err := GetClubMember(clubid, uid)
		if err != nil {
			log.Err("redis get club member err: %v", err)
		}
		if mcm == nil {
			continue
		}
		mcm.Online = cacheuser.GetUserOnlineStatus(uid)
		rooms = append(rooms, mcm)
	}
	return rooms
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
		return errors.Internal(fmt.Sprintf("delete all clubSrv error,key:%s", key), err)
	}
	return nil
}

func SetAllClub(mclubs []*mdclub.Club) error {
	for _, mClub := range mclubs {
		key := ClubHKey(mClub.ClubID)
		f := func(tx *redis.Tx) error {
			tx.Pipelined(func(p *redis.Pipeline) error {
				tx.Del(key)
				b, _ := json.Marshal(mClub)
				tx.HSet(key, "club", string(b))
				return nil
			})
			return nil
		}

		if err := cache.KV().Watch(f, key); err != nil {
			return errors.Internal("set club member failed", err)
		}
	}
	log.Err("redis reset all club list")
	return nil
}

func SetAllClubMember(mcms []*mdclub.ClubMember) error {
	err := DeleteAll(ClubMemberHKeySearch())
	if err != nil {
		return err
	}
	for _, mCm := range mcms {
		SetClubMember(mCm)
	}
	return nil
}
