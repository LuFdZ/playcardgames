package room

import (
	"math/rand"
	"playcards/model/bill"
	"encoding/json"
	enumbill "playcards/model/bill/enum"
	mbill "playcards/model/bill/mod"
	"playcards/model/club"
	mdpage "playcards/model/page"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	errroom "playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	cacheuser "playcards/model/user/cache"
	mduser "playcards/model/user/mod"
	pbroom "playcards/proto/room"
	enumcom "playcards/model/common/enum"
	cachelog "playcards/model/log/cache"
	"playcards/utils/errors"
	//"playcards/model/mail"
	"playcards/model/config"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"fmt"
	"math"
	"bytes"
)

func GenerateRangeNum(min, max int) string {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	return strconv.Itoa(randNum)
}

func GetRoomByUserID(uid int32) (*mdroom.Room, error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func RenewalRoom(pwd string, mduser *mduser.User) (int32, []int32, *mdroom.Room, error) {
	mr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return 0, nil, nil, err
	}
	if mr.ClubID > 0 {
		return 0, nil, nil, errroom.ErrClubCantRenewal
	}
	if mr.RoomType == enumroom.RoomTypeAgent {
		return 0, nil, nil, errroom.ErrRoomType
	}
	if mr.Status == enumroom.RoomStatusInit {
		return 0, nil, nil, errroom.ErrRenewalRoon
	}
	if mr.Status != enumroom.RoomStatusDelay {
		return 0, nil, nil, errroom.ErrRoomType
	}

	var ids []int32
	for _, id := range mr.Ids {
		if id != mduser.UserID {
			ids = append(ids, id)
		}
	}
	oldID := mr.RoomID
	newRoom, err := CreateRoom(mr.RoomType, mr.SubRoomType, mr.GameType, mr.StartMaxNumber,
		mr.RoundNumber, mr.GameParam, mr.SettingParam, mduser, pwd, mr.VipRoomSettingID, 0, mr.RoomAdvanceOptions) //, mr.JoinType
	if err != nil {
		log.Err("room Renewal create new room failed,%v | %v\n", mr, err)
		return 0, nil, nil, err
	}
	cacheroom.DeleteRoom(mr)
	err = cacheroom.UpdateRoom(newRoom)
	if err != nil {
		log.Err("room Renewal set redis failed,%v | %v\n", mr, err)
		return 0, nil, nil, err
	}
	return oldID, ids, newRoom, nil
}

func CreateRoom(rtype int32, srtype int32, gtype int32, maxNum int32, roundNum int32,
	gParam string, setting string, user *mduser.User, pwd string, vipRoomSettingID int32, clubID int32, RoomAdvanceOptions []string) (*mdroom.Room, //, joinType int32
	error) {
	//clubID := user.ClubID
	var err error
	hasRoom := cacheroom.ExistRoomUser(user.UserID)
	if hasRoom {
		return nil, errroom.ErrUserAlreadyInRoom
	}
	if rtype == enumroom.RoomTypeClub && clubID == 0 {
		return nil, errroom.ErrNotClubMember
	}

	if rtype == enumroom.RoomTypeClub && srtype > 0 && len(setting) == 0 {
		return nil, errroom.ErrSettingParam
	}

	//if joinType == 0 {
	//	joinType = enumroom.CanJoin
	//}

	maxNum, roundNum, gtype, gParam, RoomAdvanceOptions, err = checkGameParam(maxNum, roundNum, gtype, gParam, RoomAdvanceOptions) //, joinType
	if err != nil {
		return nil, err
	}
	//****************************************
	//roundNum = 2
	users := []*mdroom.RoomUser{}
	ids := []int32{}
	if rtype == 0 {
		rtype = enumroom.RoomTypeNom
	}
	channel := user.Channel
	version := user.Version
	mobileOs := user.MobileOs
	payerID := user.UserID
	settingParam := setting
	if rtype == enumroom.RoomTypeClub {
		mdclub, err := club.GetClubFromDB(clubID)
		if err != nil {
			return nil, err
		}
		_, creater := cacheuser.GetUserByID(mdclub.CreatorID)
		if creater == nil {
			return nil, errroom.ErrClubCreaterNotFind
		}
		//log.Debug("CreateRoomUserCreateRoom:%d|%d",vipRoomSettingID,mdclub.Setting.UserCreateRoom)
		if vipRoomSettingID == 0 && mdclub.Setting.UserCreateRoom == enumroom.UserCreateRoomClose && user.UserID != creater.UserID {
			return nil, errroom.ErrUserCreateRoom
		}

		channel = creater.Channel
		version = creater.Version
		mobileOs = creater.MobileOs
		payerID = mdclub.CreatorID
		mdcm, err := club.GetClubMember(clubID, user.UserID)
		if err != nil {
			return nil, err
		}
		if mdcm.Status != 1 {
			return nil, errroom.ErrCanNotIntoClubRoom
		}
		if srtype == enumroom.SubTypeClubMatch {
			if mdcm.ClubCoin < mdclub.Setting.ClubCoinBaseScore {
				mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
				mderr.Detail = fmt.Sprintf(mderr.Detail, mdclub.Setting.ClubCoinBaseScore)
				return nil, mderr
			}
			if mdcm.ClubCoin > enumroom.UserClubCoinLimit {
				return nil, errroom.ErrClubCoinOverdose
			}
			var mdSp *mdroom.SettingParam
			if err := json.Unmarshal([]byte(setting), &mdSp); err != nil {
				log.Err("room check thirteen clean unmarshal room param failed, %v", err)
				return nil, errroom.ErrGameParam
			}
			if mdSp.ClubCoinRate != 0 && mdSp.ClubCoinRate != 1 && mdSp.ClubCoinRate != 2 && mdSp.ClubCoinRate != 5 && mdSp.ClubCoinRate != 10 {
				return nil, errroom.ErrGameParam
			}
			mdSp.CostType = mdclub.Setting.CostType
			mdSp.CostValue = mdclub.Setting.CostValue
			mdSp.ClubCoinBaseScore = mdclub.Setting.ClubCoinBaseScore
			mdSp.CostRange = mdclub.Setting.CostRange
			mdSp.CostBase = mdclub.Setting.CostBase
			data, _ := json.Marshal(&mdSp)
			settingParam = string(data)
		}
	}
	cost := getRoomCost(gtype, maxNum, roundNum, channel, version, mobileOs, rtype)
	if rtype == enumroom.RoomTypeNom {
		roomUser := GetRoomUser(user.UserID, enumroom.UserReady, 1,
			enumroom.UserRoleMaster)
		roomUser.UserRole = enumroom.UserRolePlayerBro
		users = append(users, roomUser)
		ids = append(ids, user.UserID)
		userBalance, err := bill.GetUserBalance(user.UserID, enumbill.TypeDiamond)
		if err != nil {
			return nil, err
		}
		if userBalance.Balance < cost {
			return nil, errroom.ErrNotEnoughDiamond
		}
	} else if rtype == enumroom.RoomTypeClub {
		roomUser := GetRoomUser(user.UserID, enumroom.UserReady, 1,
			enumroom.UserRoleMaster)
		roomUser.UserRole = enumroom.UserRolePlayerBro
		users = append(users, roomUser)
		ids = append(ids, user.UserID)
		mdclub, err := club.GetClubFromDB(clubID)
		if err != nil {
			return nil, err
		}
		if mdclub.Diamond < cost {
			return nil, errroom.ErrClubNotEnoughDiamond
		}
		mdcm, err := club.GetClubMember(clubID, user.UserID)
		if err != nil {
			return nil, err
		}
		roomUser.ClubCoin = mdcm.ClubCoin
	} else if rtype == enumroom.RoomTypeAgent {
		f := func(r *mdroom.Room) bool {
			if r.Status < enumroom.RoomStatusStarted {
				return true
			}
			return false
		}
		_, _, total := cacheroom.PageAgentRoom(user.UserID, enumroom.AgentRoomAllGameType, enumroom.AgentRoomAllPage, f)
		if total >= enumroom.AgentRoomLimit {
			return nil, errroom.ErrRoomAgentLimit
		}
		userBalance, err := bill.GetUserBalance(user.UserID, enumbill.TypeDiamond)
		if err != nil {
			return nil, err
		}
		if userBalance.Balance < cost {
			return nil, errroom.ErrNotEnoughDiamond
		}
	}
	if len(pwd) == 0 {
		pwd, err = getPassWord()
		if err != nil {
			return nil, err
		}
	}
	now := gorm.NowFunc()
	mr := &mdroom.Room{
		Password:           pwd,
		GameType:           gtype,
		MaxNumber:          maxNum,
		RoundNumber:        roundNum,
		RoundNow:           1,
		GameParam:          gParam,
		Status:             enumroom.RoomStatusInit,
		Giveup:             enumroom.NoGiveUp,
		Users:              users,
		RoomType:           rtype,
		PayerID:            payerID,
		GiveupAt:           &now,
		Ids:                ids,
		Cost:               cost,
		StartMaxNumber:     maxNum,
		CostType:           enumroom.CostTypeDiamond,
		Flag:               enumroom.RoomNoFlag,
		BankerList:         []int32{user.UserID},
		SubRoomType:        srtype,
		SettingParam:       settingParam,
		VipRoomSettingID:   vipRoomSettingID,
		RoomAdvanceOptions: RoomAdvanceOptions,
		ReadyUserMap:       make(map[int32]*mdroom.RoomUser),
		ReadyDate:          &now,
	}
	if mr.RoomType == enumroom.RoomTypeClub {
		mr.ClubID = clubID
	}

	f := func(tx *gorm.DB) error {
		err := dbroom.CreateRoom(tx, mr)
		if err != nil {
			log.Err("room create failed,%v | %v", mr, err)
			return err
		}

		//err = RoomCreateBalance(mr, user.UserID)
		err = cacheroom.SetRoom(mr)
		if err != nil {
			log.Err("room create set redis failed,%v | %v\n", mr, err)
			return err
		}
		if rtype == enumroom.RoomTypeNom || rtype == enumroom.RoomTypeClub {
			err = cacheroom.SetRoomUser(mr.RoomID, mr.Password, user.UserID)
			if err != nil {
				log.Err("room create set room user redis failed,%v | %v\n", mr, err)
				return err
			}
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func CheckUserInClub(uid int32, clubid int32) error {
	_, err := club.GetClubMember(clubid, uid)
	if err != nil {
		return err
	}
	return nil
}

func getPassWord() (string, error) {
	pwdNew := GenerateRangeNum(enumroom.RoomCodeMin, enumroom.RoomCodeMax)
	exist := cacheroom.CheckRoomExist(pwdNew)
	for i := 0; exist && i < 3; i++ {
		pwdNew = GenerateRangeNum(enumroom.RoomCodeMin, enumroom.RoomCodeMax)
		exist = cacheroom.CheckRoomExist(pwdNew)
	}
	if exist {
		return "", errroom.ErrRoomPwdExisted
	}
	return pwdNew, nil
}

func getRoomCost(gType int32, maxNumber int32, roundNumber int32, channel string, version string, mobileOs string, roomtype int32) int64 {
	var diamond int64
	var cost int32 = 1
	var configID string

	//if roomtype == enumroom.RoomTypeNom {
	//	configID = enumroom.ConsumeOpen
	//} else if roomtype == enumroom.RoomTypeAgent {
	//	configID = enumroom.AgentConsumeOpen
	//} else if roomtype == enumroom.RoomTypeClub {
	//	configID = enumroom.ClubConsumeOpen
	//}
	//if gType == enumroom.ThirteenGameType {
	//	cost = enumroom.ThirteenGameCost
	//} else if gType == enumroom.NiuniuGameType {
	//	cost = enumroom.NiuniuGameCost
	//} else if gType == enumroom.DoudizhuGameType {
	//	cost = enumroom.DoudizhuGameCost
	//} else if gType == enumroom.FourCardGameType {
	//	cost = enumroom.FourcardGameCost
	//} else if gType == enumroom.TwoCardGameType {
	//	cost = enumroom.TwocardGameCost
	//}

	if gType == enumroom.ThirteenGameType {
		configID = enumroom.ThirteenConsume
	} else if gType == enumroom.NiuniuGameType {
		configID = enumroom.NiuniuConsume
	} else if gType == enumroom.DoudizhuGameType {
		configID = enumroom.DoudizhuConsume
	} else if gType == enumroom.FourCardGameType {
		configID = enumroom.FourcardConsume
	} else if gType == enumroom.TwoCardGameType {
		configID = enumroom.TwocardConsume
	} else if gType == enumroom.RunCardGameType {
		configID = enumroom.RuncardConsume
	}

	diamond = int64(maxNumber * roundNumber * cost)

	rate := config.CheckConfigCondition(configID, channel, version, mobileOs)
	diamond = int64(rate * float64(diamond))
	//log.Debug("CreateRoom GetRoomCost:%d|%d,Condition:%s|%s|%s\n",diamond,rate,channel,version,mobileOs)
	return diamond
}

//func getRoomJournalType(gameType int32) int32 {
//	var jType int32
//	if gameType == enumroom.ThirteenGameType {
//		jType = enumbill.JournalTypeThirteenFreeze
//	} else if gameType == enumroom.NiuniuGameType {
//		jType = enumbill.JournalTypeNiuniuFreeze
//	} else if gameType == enumroom.DoudizhuGameType {
//		jType = enumbill.JournalTypeDoudizhuFreeze
//	}
//	return jType
//}

func WatchRoom(pwd string, mduser *mduser.User) (*mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(mduser.UserID)
	if hasRoom {
		return nil, errroom.ErrUserAlreadyInRoom
	}
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}

	if mdr.RoomAdvanceOptions[0] != enumroom.CanJoin && int32(len(mdr.Users)+len(mdr.WatchIds)+len(mdr.ReadyUserMap))+1 > mdr.StartMaxNumber {
		return nil, errroom.ErrRoomFull
	}

	if mdr.RoomAdvanceOptions[0] != enumroom.CanJoin && (mdr.RoundNow > 1 ||
		(mdr.RoundNow == 1 && mdr.Status > enumroom.RoomStatusAllReady)) {
		return nil, errroom.ErrCanNotJoin
	}

	hasRoom = false
	for _, watchid := range mdr.WatchIds {
		if watchid == mduser.UserID {
			hasRoom = true
			break
		}
	}
	if hasRoom {
		return nil, errroom.ErrUserAlreadyInRoom
	}
	err = CheckUserInClub(mduser.UserID, mdr.ClubID)
	if mdr.ClubID > 0 && err != nil {
		return nil, errroom.ErrNotClubMember
	}

	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	//fmt.Printf("WatchRoomLimit:%d|%d\n", len(mdr.WatchIds), enumroom.WatchLimit)
	//if len(mdr.WatchIds)+1 > enumroom.WatchLimit {
	//	return nil, errroom.ErrWatchLimit
	//}

	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, mduser.UserID)
		if mdcm.Status != 1 {
			return nil, errroom.ErrCanNotIntoClubRoom
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return nil, mderr
		}
	}
	mdr.WatchIds = append(mdr.WatchIds, mduser.UserID)
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mduser.UserID)
	if err != nil {
		log.Err("room user join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	return mdr, nil
}

func AutoSetReady(roomUser *mdroom.RoomUser, mdr *mdroom.Room) (*mdroom.Room, error) { //uid int32, mdr *mdroom.Room
	//mdr, err := cacheroom.GetRoomUserID(uid)
	//if err != nil {
	//	return nil,err
	//}
	//var num int32 = int32(len(mdr.Users))
	//if num >= mdr.StartMaxNumber {
	//	return nil,errroom.ErrRoomFull
	//

	//p := 0
	//positionArray := make([]int32, mdr.StartMaxNumber)
	//for n := 0; n < len(positionArray); n++ {
	//	positionArray[n] = 0
	//}
	//for _, ru := range mdr.Users {
	//	if ru.UserID == uid {
	//		return nil, errroom.ErrUserAlreadyInRoom
	//	}
	//	positionArray[ru.Position-1] = 1
	//}
	//for n := 0; n < len(positionArray); n++ {
	//	if positionArray[n] == 0 {
	//		p = n
	//	}
	//}
	//roomUser := GetRoomUser(uid, enumroom.UserUnready, int32(p+1),
	//	enumroom.UserRoleSlave)
	//delete(mdr.ReadyUserMap, uid)

	//mdr.ReadyUserMap[uid] = roomUser
	//mdr.Ids = append(mdr.Ids, uid)
	//err = cacheroom.UpdateRoom(mdr)
	//if err != nil {
	//	log.Err("room join set session failed, %v|%v\n", err, mdr)
	//	return nil,err
	//}
	//fmt.Printf("AutoSetReadyAAA:%+v|%+v\n",mdr.Users,mdr.Ids)
	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, roomUser.UserID)
		if mdcm.Status != 1 {
			return nil, errroom.ErrCanNotIntoClubRoom
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return nil, mderr
		}
		roomUser.ClubCoin = mdcm.ClubCoin
	}
	roomUser.UserRole = enumroom.UserRolePlayerBro
	roomUser.Ready = enumroom.UserUnready
	mdr.Users = append(mdr.Users, roomUser)
	mdr.Ids = append(mdr.Ids, roomUser.UserID)
	mdr.MaxNumber ++
	//fmt.Printf("AutoSetReadyBBB:%+v\n",mdr.Users)
	//err = cacheroom.UpdateRoom(mdr)
	//if err != nil {
	//	log.Err("room join set session failed, %v|%v\n", err, mdr)
	//	return nil,err
	//}
	//UpdateRoom(mdr)
	return mdr, nil
}

func SitDown(pwd string, uid int32) (bool, *mdroom.RoomUser, *mdroom.Room, error) { //mdu *mduser.User
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return false, nil, nil, err
	}
	if pwd != mdr.Password {
		return false, nil, nil, errroom.ErrUserNotInRoom
	}
	//fmt.Printf("AAAAAAASitDown:%s|%d|%d|%t\n",mdr.RoomAdvanceOptions[0], mdr.Status,mdr.RoundNow,
	//	mdr.RoundNow == 1 &&  mdr.Status > enumroom.RoomStatusStarted)
	// && mdr.Status > enumroom.RoomStatusAllReady

	//if mdr.RoomAdvanceOptions[0] != enumroom.CanJoin && (mdr.RoundNow > 1 ||
	//	(mdr.RoundNow == 1 && mdr.Status > enumroom.RoomStatusAllReady)) {
	//	return false, nil, nil, errroom.ErrCanNotJoin
	//}
	if mdr.ClubID > 0 {
		err = CheckUserInClub(uid, mdr.ClubID)
		if err != nil {
			return false, nil, nil, errroom.ErrNotClubMember
		}
	}

	if mdr.Giveup == enumroom.WaitGiveUp {
		return false, nil, nil, errroom.ErrInGiveUp
	}

	var num int32 = int32(len(mdr.Users))
	num += int32(len(mdr.ReadyUserMap))
	//fmt.Printf("AAAAAAAASitDownStartMaxNumber:%d|%d|%d|%d\n", num, int32(len(mdr.Users)), int32(len(mdr.ReadyUserMap)), mdr.StartMaxNumber)
	if num >= mdr.StartMaxNumber {
		return false, nil, nil, errroom.ErrRoomFull
	}
	p := 0
	positionArray := make([]int32, mdr.StartMaxNumber)
	for n := 0; n < len(positionArray); n++ {
		positionArray[n] = 0
	}
	var allUser []*mdroom.RoomUser
	for _, ru := range mdr.Users {
		allUser = append(allUser, ru)
	}
	for _, ru := range mdr.ReadyUserMap {
		allUser = append(allUser, ru)
	}
	for _, ru := range allUser {
		if ru.UserID == uid {
			return false, nil, nil, errroom.ErrUserAlreadyInRoom
		}
		positionArray[ru.Position-1] = 1
		//positionArray[ru.Position-1] = 1
	}

	for n := 0; n < len(positionArray); n++ {
		if positionArray[n] == 0 {
			p = n
			break
		}
	}
	ready := enumroom.UserUnready
	if mdr.Status == enumroom.RoomStatusInit && mdr.RoundNow == 1 {
		ready = enumroom.UserReady
	}
	roomUser := GetRoomUser(uid, ready, int32(p+1), //int32(p+1)
		enumroom.UserRoleSlave)

	var wids []int32
	inWatchList := false
	for _, watchid := range mdr.WatchIds {
		if watchid == uid {
			inWatchList = true
		} else {
			wids = append(wids, watchid)
		}
	}
	if !inWatchList {
		return false, nil, nil, errroom.ErrNotInWatchList
	}
	mdr.WatchIds = wids
	//sub := time.Now().Sub(*mdr.ReadyDate)
	//if sub.Seconds() < 2 {
	//	t1 := mdr.ReadyDate
	//	s, _ := time.ParseDuration("2s")
	//	t2 := t1.Add(s)
	//	mdr.ReadyDate = &t2
	//}
	t1 := mdr.ReadyDate
	s, _ := time.ParseDuration("2s")
	t2 := t1.Add(s)
	mdr.ReadyDate = &t2

	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, uid)
		if mdcm.Status != 1 {
			return false, nil, nil, errroom.ErrCanNotIntoClubRoom
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return false, nil, nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin > enumroom.UserClubCoinLimit {
			return false, nil, nil, errroom.ErrClubCoinOverdose
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return false, nil, nil, mderr
		}
		roomUser.ClubCoin = mdcm.ClubCoin
	}
	roomUser.UserRole = enumroom.UserRolePlayerBro
	roomUser.UserSitDownRound = mdr.RoundNow
	//log.Debug("SitDownCreatePlayerRoom:%d|%d|%d|%+v", mdr.RoundNow, mdr.Status, uid, mdr)
	if mdr.RoundNow > 1 || (mdr.RoundNow == 1 && mdr.Status > enumroom.RoomStatusInit ) {
		f := func(tx *gorm.DB) error {
			pr := &mdroom.PlayerRoom{
				UserID:    uid,
				RoomID:    mdr.RoomID,
				GameType:  mdr.GameType,
				PlayTimes: 0,
			}
			err := dbroom.CreatePlayerRoom(tx, pr)
			if err != nil {
				log.Err("niuniu create player room err:%v|%v\n", pr, err)
				return err
			}
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			cachelog.SetErrLog(enumroom.ServiceCode, err.Error())
			return false, nil, nil, err
		}
	}

	if mdr.Status != enumroom.RoomStatusInit {
		if mdr.Status > enumroom.RoomStatusReInit || mdr.RoundNow == mdr.RoundNumber {
			return false, nil, nil, errroom.ErrGameIsDone
		}
		if _, ok := mdr.ReadyUserMap[uid]; ok {
			return false, nil, nil, errroom.ErrAlreadyReady
		}
		roomUser.UserRole = enumroom.UserRoleWatchBro

		mdr.ReadyUserMap[uid] = roomUser
		//mdr.Ids = append(mdr.Ids, uid)
		err = cacheroom.UpdateRoom(mdr)
		if err != nil {
			log.Err("room join set session failed, %v|%v\n", err, mdr)
			return false, nil, nil, err
		}
		return false, roomUser, mdr, nil
		//return false, nil, nil, errroom.ErrNotReadyStatus
	}
	allReady := false
	if mdr.RoomType == enumroom.RoomTypeAgent && len(mdr.Users) == 0 {
		roomUser.Role = enumroom.UserRoleMaster
		if mdr.GameType == enumroom.FourCardGameType || mdr.GameType == enumroom.TwoCardGameType {
			mdr.BankerList[0] = roomUser.UserID
		}
	}
	if mdr.Status == enumroom.RoomStatusInit {
		mdr.Users = append(mdr.Users, roomUser)
		mdr.Ids = append(mdr.Ids, uid)
		if mdr.MaxNumber < mdr.StartMaxNumber {
			mdr.MaxNumber ++
		}

		num = 0
		//allReady := true
		for _, user := range mdr.Users {
			if user.UserRole == enumroom.UserRolePlayerBro {
				if user.Ready != enumroom.UserReady { //allReady &&
					allReady = false
				} else {
					num ++
				}
			}
		}
		mdr.Users[0].Role = enumroom.UserRoleMaster
		//if allReady && num == mdr.MaxNumber { // mdr.MaxNumber
		//	if mdr.RoundNow == 1 {
		//		mdr.Users[0].Role = enumroom.UserRoleMaster
		//	}
		//	mdr.Status = enumroom.RoomStatusAllReady
		//}
		//if allReady && num == mdr.MaxNumber && mdr.RoundNow == 1 {
		//	err = RoomCreateBalance(mdr, mdr.PayerID)
		//	if err != nil {
		//		log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, mdr.PayerID, err)
		//		return false, nil, nil, err
		//	}
		//}
	}
	//mdr.MaxNumber ++

	//allReady := true
	//t := time.Now()
	//num = 0
	//for _, user := range mdr.Users {
	//	num++
	//	if user.UserID == uid {
	//		if user.Ready == enumroom.UserReady {
	//			return false, nil, nil, errroom.ErrAlreadyReady
	//		}
	//		user.Ready = enumroom.UserReady
	//		user.UpdatedAt = &t
	//	}
	//	if allReady && user.Ready != enumroom.UserReady {
	//		allReady = false
	//	}
	//}
	//if allReady && num == mdr.MaxNumber {
	//	if mdr.RoundNow == 1 {
	//		mdr.Users[0].Role = enumroom.UserRoleMaster
	//	}
	//	if mdr.Shuffle > 0 {
	//		mdr.Status = enumroom.RoomStatusShuffle
	//		mdr.ShuffleAt = &t
	//	} else {
	//		mdr.Status = enumroom.RoomStatusAllReady
	//	}
	//}

	if allReady && num == mdr.MaxNumber && mdr.RoundNow == 1 {
		err = RoomCreateBalance(mdr, mdr.PayerID)
		if err != nil {
			log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, mdr.PayerID, err)
			return false, nil, nil, err
		}
		//err = roomBackUnFreezeAndBalance(mdr)
		//if err != nil {
		//	log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
		//	return false, nil, err
		//}
	}

	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room join set session failed, %v|%v\n", err, mdr)
		return false, nil, nil, err
	}

	UpdateRoom(mdr)
	return allReady, roomUser, mdr, nil //allReady, roomUser, mdr, nil
}

func JoinRoom(pwd string, mduser *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	//hasRoom := false
	//for _, watchid := range mdr.WatchIds {
	//	if watchid == mduser.UserID {
	//		hasRoom = true
	//		break
	//	}
	//}

	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}

	//if !hasRoom {
	//
	//}
	hasRoom := cacheroom.ExistRoomUser(mduser.UserID)
	if hasRoom {
		return nil, nil, errroom.ErrUserAlreadyInRoom
	}
	if mdr.ClubID > 0 {
		err = CheckUserInClub(mduser.UserID, mdr.ClubID)
		if err != nil {
			return nil, nil, errroom.ErrNotClubMember
		}
	}

	if mdr.Status != enumroom.RoomStatusInit {
		return nil, nil, errroom.ErrNotReadyStatus
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, nil, errroom.ErrInGiveUp
	}

	num := len(mdr.Users)
	if num >= (int)(mdr.MaxNumber) {
		return nil, nil, errroom.ErrRoomFull
	}
	p := 0
	positionArray := make([]int32, mdr.MaxNumber)
	for n := 0; n < len(positionArray); n++ {
		positionArray[n] = 0
	}
	for _, ru := range mdr.Users {
		if ru.UserID == mduser.UserID {
			return nil, nil, errroom.ErrUserAlreadyInRoom
		}
		positionArray[ru.Position-1] = 1
	}
	for n := 0; n < len(positionArray); n++ {
		if positionArray[n] == 0 {
			p = n
		}
	}
	roomUser := GetRoomUser(mduser.UserID, enumroom.UserUnready, int32(p+1),
		enumroom.UserRoleSlave)
	if mdr.RoomType == enumroom.RoomTypeAgent && len(mdr.Users) == 0 {
		roomUser.Role = enumroom.UserRoleMaster
		if mdr.GameType == enumroom.FourCardGameType || mdr.GameType == enumroom.TwoCardGameType {
			mdr.BankerList[0] = roomUser.UserID
		}
	}

	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, mduser.UserID)
		if mdcm.Status != 1 {
			return nil, nil, errroom.ErrCanNotIntoClubRoom
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return nil, nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return nil, nil, mderr
		}
		roomUser.ClubCoin = mdcm.ClubCoin
	}

	mdr.Users = append(mdr.Users, roomUser)
	mdr.Ids = append(mdr.Ids, mduser.UserID)
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room join set session failed, %v|%v\n", err, mdr)
		return nil, nil, err
	}
	err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mduser.UserID)
	if err != nil {
		log.Err("room user join set session failed, %v|%v\n", err, mdr)
		return nil, nil, err
	}
	UpdateRoom(mdr)
	return roomUser, mdr, nil
}

func GetReady(pwd string, uid int32, shuffle bool) (bool, *mdroom.Room, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return false, nil, err
	}
	if pwd != mdr.Password {
		return false, nil, errroom.ErrUserNotInRoom
	}

	//if mdr.JoinType != enumroom.CanJoin || mdr.RoundNow == mdr.RoundNumber {
	//	return false, nil, errroom.ErrCanNotJoinOrGameOver
	//}
	if mdr.RoomAdvanceOptions[0] != enumroom.CanJoin && mdr.Status > enumroom.RoomStatusInit {
		return false, nil, errroom.ErrNotReadyStatus
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return false, nil, errroom.ErrInGiveUp
	}
	ru := mdr.GetRoomUser(uid)
	if ru == nil {
		return false, nil, errroom.ErrNotInGame
	}
	if ru.UserRole != enumroom.UserRolePlayerBro {
		return false, nil, errroom.ErrNotInGame
	}
	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdclub, err := club.GetClubInfo(mdr.ClubID)
		if err != nil {
			return false, nil, err
		}
		mdcm, err := club.GetClubMember(mdr.ClubID, uid)
		if err != nil {
			return false, nil, errroom.ErrNotClubMember
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check setting param clean unmarshal room param failed, %v", err)
			return false, nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, mdclub.Setting.ClubCoinBaseScore)
			return false, nil, mderr
		}
	}

	//sub := time.Now().Sub(*mdr.ReadyDate)
	//if sub.Seconds() < 2 {
	//	t1 := mdr.ReadyDate
	//	s, _ := time.ParseDuration("2s")
	//	t2 := t1.Add(s)
	//	mdr.ReadyDate = &t2
	//}
	t1 := mdr.ReadyDate
	s, _ := time.ParseDuration("2s")
	t2 := t1.Add(s)
	mdr.ReadyDate = &t2

	t := time.Now()
	if mdr.Status != enumroom.RoomStatusInit {
		for _, user := range mdr.Users {
			if user.UserID == uid {
				if user.Ready == enumroom.UserReady {
					return false, nil, errroom.ErrAlreadyReady
				}
				user.Ready = enumroom.UserReady
				user.UpdatedAt = &t
				err = cacheroom.UpdateRoom(mdr)
				if err != nil {
					log.Err("room ready failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
					return false, nil, err
				}
				break
			}
		}
		return false, mdr, nil
	}
	if shuffle {
		mdr.Shuffle = uid
	}
	allReady := true

	var num int32
	for _, user := range mdr.Users {
		if user.UserRole != enumroom.UserRolePlayerBro {
			continue
		}
		num++
		if user.UserID == uid {
			if user.Ready == enumroom.UserReady {
				return false, nil, errroom.ErrAlreadyReady
			}
			user.Ready = enumroom.UserReady
			user.UpdatedAt = &t
		}
		if allReady && user.Ready != enumroom.UserReady {
			allReady = false
		}
	}
	if allReady && num >= enumroom.GameNumberLimit[mdr.GameType][0] { //  num == mdr.MaxNumber
		if mdr.RoundNow == 1 {
			mdr.Users[0].Role = enumroom.UserRoleMaster
		}
		if mdr.Shuffle > 0 {
			mdr.Status = enumroom.RoomStatusShuffle
			mdr.ShuffleAt = &t
		} else {
			mdr.Status = enumroom.RoomStatusAllReady
		}
	}

	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room ready failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
		return false, nil, err
	}
	//fmt.Printf("AAAAAAAA:%t|%d|%d|%d\n", allReady, num, mdr.MaxNumber, mdr.RoundNow)
	//if allReady && num == mdr.MaxNumber && mdr.RoundNow == 1 {
	//	err = RoomCreateBalance(mdr, mdr.PayerID)
	//	if err != nil {
	//		log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
	//		return false, nil, err
	//	}
	//}
	if mdr.RoundNow == 1 {
		allReady = false
	}

	return allReady, mdr, nil
}

func LeaveRoom(mduser *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	mr, err := cacheroom.GetRoomUserID(mduser.UserID)
	if err != nil {
		return nil, nil, err
	}

	for _, ru := range mr.Users {
		if ru.UserID == mduser.UserID && ru.UserRole == enumroom.UserRoleSuspendBro {
			return nil, nil, errroom.ErrNotInGame
		}
	}
	isWatch := false
	var wids []int32
	for _, uid := range mr.WatchIds {
		if uid == mduser.UserID {
			isWatch = true
		} else {
			wids = append(wids, uid)
		}
	}
	//fmt.Printf("BBBBBBBBBLeaveRoomisWatch:%t|%+v\n", isWatch, mr.WatchIds)
	if isWatch {
		mr.WatchIds = wids
		err = cacheroom.DeleteRoomUser(mduser.UserID)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room join set session failed, %v|%v\n", err, mr)
			return nil, nil, err
		}
		return nil, mr, err
	}
	isReady := false
	rmap := make(map[int32]*mdroom.RoomUser)
	var ReadyRoomUser *mdroom.RoomUser
	var rids []int32
	for uid, ru := range mr.ReadyUserMap {
		if uid == mduser.UserID {
			isReady = true
			ReadyRoomUser = ru
		} else {
			rmap[uid] = ru
			rids = append(rids, uid)
		}
	}
	if isReady {
		mr.ReadyUserMap = rmap
		mr.Ids = rids
		err = cacheroom.DeleteRoomUser(mduser.UserID)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room join set session failed, %v|%v\n", err, mr)
			return nil, nil, err
		}
		return ReadyRoomUser, mr, err
	}

	if mr.RoundNow > 1 || (mr.RoundNow == 1 && mr.Status > enumroom.RoomStatusInit) {
		return nil, nil, errroom.ErrGameHasBegin
	}
	if mr.Giveup == enumroom.WaitGiveUp {
		return nil, nil, errroom.ErrInGiveUp
	}

	newUsers := []*mdroom.RoomUser{}
	roomUser := &mdroom.RoomUser{}
	handle := 0
	masterLeave := false
	for _, u := range mr.Users {
		if u.UserID == mduser.UserID {
			handle = 1
			roomUser = u
			if u.Role == enumroom.UserRoleMaster &&
				(mr.RoomType == enumroom.RoomTypeNom) {
				log.Info("delete room cause master leave.user:%d,room:%d\n",
					mduser.UserID, mr.RoomID)
				masterLeave = true
			}
		} else {
			newUsers = append(newUsers, u)
		}
	}
	mr.Users = newUsers

	if mr.GameType == enumroom.FourCardGameType || mr.GameType == enumroom.TwoCardGameType {
		var blist []int32
		for _, bid := range mr.BankerList {
			if bid != mduser.UserID {
				blist = append(blist, bid)
			}
		}
		if len(blist) == 0 && len(mr.Users) > 0 {
			blist = append(blist, mr.Users[0].UserID)
		}
		mr.BankerList = blist
	}
	var ids []int32
	for _, user := range mr.Users {
		ids = append(ids, user.UserID)
	}
	mr.Ids = ids
	if handle == 0 {
		return nil, nil, errroom.ErrUserNotInRoom
	}
	if roomUser.Role == enumroom.UserRoleMaster && mr.RoomType != enumroom.RoomTypeNom && len(mr.Users) > 0 {
		mr.Users[0].Role = enumroom.UserRoleMaster
	}
	//RoomTypeNom 普通开房解散条件 人员全部退出 || 房主退出
	//RoomTypeAgent 代开房解散条件 代开房者主动解散
	//RoomTypeClub 俱乐部开房解散条件 人员全部退出
	if (len(newUsers) == 0 && mr.RoomType != enumroom.RoomTypeAgent) ||
		(masterLeave && mr.RoomType == enumroom.RoomTypeNom) {
		mr.Users = nil
		mr.Status = enumroom.RoomStatusDestroy
		err := cacheroom.DeleteAllRoomUser(mr.Password, "LeaveRoom")
		if err != nil {
			log.Err("leave room delete all users failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacheroom.DeleteRoom(mr)
		if err != nil {
			log.Err("leave room delete room failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}

		err = RoomRefund(mr)
		if err != nil {
			log.Err("leave room refund failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
	} else {
		err := cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room leave room delete failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacheroom.DeleteRoomUser(mduser.UserID)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		if mr.RoomType != enumroom.RoomTypeNom && len(newUsers) > 0 {
			newUsers[0].Role = enumroom.UserRoleMaster
		}
		//UpdateRoom(mr)
	}
	UpdateRoom(mr)
	return roomUser, mr, nil
}

func GiveUpGame(pwd string, uid int32) ([]int32, *mdroom.GiveUpGameResult, *mdroom.Room,
	error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, nil, nil, err
	}

	ru := mr.GetRoomUser(uid)
	if ru != nil {
		if ru.UserID == uid && ru.UserRole != enumroom.UserRolePlayerBro {
			return nil, nil, nil, errroom.ErrNotInGame
		}
	}

	if (mr.RoundNumber == 1 && mr.Status < enumroom.RoomStatusStarted) || mr.RoundNumber == mr.RoundNow {
		return nil, nil, nil, errroom.ErrNotReadyStatus
	}

	if mr.Giveup == enumroom.WaitGiveUp {
		return GiveUpVote(pwd, 1, uid)
	}

	isWatch := false
	var wids []int32
	for _, wuid := range mr.WatchIds {
		if uid == wuid {
			isWatch = true
		} else {
			wids = append(wids, uid)
		}
	}
	//fmt.Printf("GiveUpGame:%t|%v|%v\n", isWatch, wids, mr.WatchIds)
	if isWatch {
		mr.WatchIds = wids
		err = cacheroom.DeleteRoomUser(uid)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, nil, err
		}
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room join set session failed, %v|%v\n", err, mr)
			return nil, nil, nil, err
		}
		return nil, nil, mr, err
	}

	giveUpResult := mr.GiveupGame
	var list []*mdroom.UserState
	agreeGiveUp := 0
	for _, user := range mr.Users {
		if user.UserRole != enumroom.UserRolePlayerBro {
			agreeGiveUp++
			continue
		}
		var state int32
		//scoketstatus := cacher.GetUserStatus(user.UserID)
		scoketstatus := cacheuser.GetUserOnlineStatus(user.UserID)
		if user.UserID == uid {
			state = enumroom.UserStateLeader
			agreeGiveUp++
		} else if scoketstatus == enumroom.SocketClose {
			state = enumroom.UserStateOffline
			//agreeGiveUp++
		} else {
			state = enumroom.UserStateWaiting
		}
		us := &mdroom.UserState{
			UserID: user.UserID,
			State:  state,
		}
		list = append(list, us)
	}
	if agreeGiveUp == len(mr.Users) {
		mr.Status = enumroom.RoomStatusGiveUp
		giveUpResult.Status = enumroom.GiveupStatusAgree
	} else {
		mr.Giveup = enumroom.WaitGiveUp
		now := gorm.NowFunc()
		mr.GiveupAt = &now
		giveUpResult.Status = enumroom.GiveupStatusWairting
	}

	giveUpResult.UserStateList = list
	mr.GiveupGame = giveUpResult
	if mr.Status == enumroom.RoomStatusGiveUp {
		RoundOverRoomClubCoin(mr)
		err = cacheroom.DeleteAllRoomUser(mr.Password, "GiveUpGame")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil, nil, nil, err
		}
		err = cacheroom.DeleteRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
		err = cacheroom.SetRoomDelete(mr.GameType, mr.RoomID)
		if err != nil {
			log.Err("give up set delete room redis err, %d|%v\n", mr.RoomID, err)
		}
		if mr.RoomType == enumroom.RoomTypeAgent {
			cacheroom.SetAgentRoom(mr)
		}
		//客户端在投票阶段断线，投票解散后重连，恢复游戏 需要将投票状态设为未投票，以显示结算结果
		mr.Giveup = enumroom.NoGiveUp
		//
		UpdateRoom(mr)
		//RoomRefund(mr)
	} else {
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	}

	log.Info("give up game:%d|%v", uid, giveUpResult)
	return mr.Ids, &giveUpResult, mr, nil
}

func GiveUpVote(pwd string, status int32, uid int32) ([]int32, *mdroom.GiveUpGameResult, *mdroom.Room,
	error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, nil, nil, err
	}
	ru := mr.GetRoomUser(uid)
	if ru != nil {
		if ru.UserID == uid && ru.UserRole != enumroom.UserRolePlayerBro {
			return nil, nil, nil, errroom.ErrNotInGame
		}
	}
	if mr.Giveup != enumroom.WaitGiveUp {
		return nil, nil, nil, errroom.ErrNotInGiveUp
	}

	giveUpResult := mr.GiveupGame
	if status == 1 {
		status = enumroom.UserStateAgree
	} else {
		status = enumroom.UserStateDisagree
	}
	giveup := enumroom.GiveupStatusWairting
	agreeGiveUp := 0
	for _, userstate := range mr.GiveupGame.UserStateList {
		if userstate.UserID == uid {
			if userstate.State != enumroom.UserStateWaiting && userstate.State != enumroom.UserStateOffline {
				return nil, nil, nil, errroom.ErrAlreadyVoted
			}
			userstate.State = status
			if userstate.State == enumroom.UserStateDisagree {
				giveup = enumroom.GiveupStatusDisAgree
				break
			}

		} else if userstate.State == enumroom.UserStateDisagree {
			giveup = enumroom.GiveupStatusDisAgree
			break
		}

		if userstate.State != enumroom.UserStateDisagree &&
			userstate.State != enumroom.UserStateWaiting &&
			userstate.State != enumroom.UserStateOffline {
			agreeGiveUp++
		}
	}
	if agreeGiveUp == len(mr.GiveupGame.UserStateList) {
		RoundOverRoomClubCoin(mr)
		mr.Status = enumroom.RoomStatusGiveUp
		giveUpResult.Status = enumroom.GiveupStatusAgree
		err = cacheroom.DeleteAllRoomUser(mr.Password, "GiveUpVote")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil, nil, nil, err
		}
		err = cacheroom.DeleteRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
		err = cacheroom.SetRoomDelete(mr.GameType, mr.RoomID)
		if err != nil {
			log.Err("give up set delete room redis err, %d|%v\n", mr.RoomID, err)
		}
		if mr.RoomType == enumroom.RoomTypeAgent {
			cacheroom.SetAgentRoom(mr)
		}
		//客户端在投票阶段断线，投票解散后重连，恢复游戏 需要将投票状态设为未投票，以显示结算结果
		mr.Giveup = enumroom.NoGiveUp
		//
		UpdateRoom(mr)
		err = RoomRefund(mr)
		if err != nil {
			log.Err("room give up refund failed, %v", err)
			//return nil, nil, nil, err
		}
	} else if giveup == enumroom.GiveupStatusDisAgree {
		mr.Giveup = enumroom.NoGiveUp
		giveUpResult.Status = enumroom.GiveupStatusDisAgree
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	} else {
		mr.Giveup = enumroom.WaitGiveUp
		giveUpResult.Status = enumroom.GiveupStatusWairting
		err = cacheroom.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	}

	log.Info("give up game vote:%d|%v", uid, giveUpResult)
	return mr.Ids, &giveUpResult, mr, nil
}

func GetRoomUser(uid int32, ready int32, position int32, //mdu *mduser.User
	role int32) *mdroom.RoomUser {
	return &mdroom.RoomUser{
		UserID: uid,
		Ready:  ready,
		//Nickname:  u.Nickname,
		Position: position,
		//Icon:      u.Icon,
		//Sex:       u.Sex,
		Role: role,
		//Location:  u.Location,
	}
}

func Shock(uid int32, sendid int32) error {
	fromMr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return err
	}

	toMr, err := cacheroom.GetRoomUserID(sendid)
	if err != nil {
		return err
	}
	if fromMr.RoomID != toMr.RoomID {
		return errroom.ErrNotInSameRoon
	}
	if fromMr.Status != enumroom.RoomStatusStarted {
		return errroom.ErrNotReadyStatus
	}
	return nil
}

func UserRoomCheck(uid int32) (*mdroom.Room, error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}
	if mr.Status > enumroom.RoomStatusDone {
		return nil, errroom.ErrGameIsDone
	}
	return mr, nil
}

//func TestClean() error {
//	cacher.FlushAll()
//	f := func(tx *gorm.DB) error {
//		dbr.DeleteAll(tx)
//		return nil
//	}
//	err := db.Transaction(f)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func PageFeedbackList(page *mdpage.PageOption, fb *mdroom.Feedback) (
	[]*mdroom.Feedback, int64, error) {
	return dbroom.PageFeedbackList(db.DB(), page, fb)
}

func CreateFeedback(fb *mdroom.Feedback, uid int32, ip string) (*mdroom.Feedback, error) {
	fb.UserID = uid
	fb.LoginIP = ip
	return dbroom.CreateFeedback(db.DB(), fb)
}

func RoomResultList(page *mdpage.PageOption, uid int32, gtype int32) (*pbroom.RoomResultListReply, error) {
	var list []*pbroom.RoomResults
	if page == nil {
		page.Page = 1
		page.PageSize = 20
	}
	rooms, rows, err := dbroom.PageRoomResultList(db.DB(), uid, gtype, page)
	if err != nil {
		return nil, err
	}
	for _, mr := range rooms {
		result := &mdroom.RoomResults{
			Status:      mr.Status,
			Password:    mr.Password,
			GameType:    mr.GameType,
			CreatedAt:   mr.CreatedAt,
			RoundNow:    mr.RoundNow,
			List:        mr.UserResults,
			SubRoomType: mr.SubRoomType,
			RoomType:    mr.RoomType,
		}
		list = append(list, result.ToProto())
	}
	out := &pbroom.RoomResultListReply{
		List:  list,
		Count: rows,
	}
	return out, nil
}

func PageClubRoomResultList(page *mdpage.PageOption, clubid int32) (*pbroom.RoomResultListReply, error) {
	var list []*pbroom.RoomResults
	if page == nil {
		page.Page = 1
		page.PageSize = 20
	}
	rooms, rows, err := dbroom.PageClubRoomResultList(db.DB(), clubid, page)
	if err != nil {
		return nil, err
	}
	for _, mr := range rooms {
		result := &mdroom.RoomResults{
			RoomID:      mr.RoomID,
			Status:      mr.Status,
			Password:    mr.Password,
			GameType:    mr.GameType,
			CreatedAt:   mr.CreatedAt,
			RoundNow:    mr.RoundNow,
			List:        mr.UserResults,
			SubRoomType: mr.SubRoomType,
			RoomType:    mr.RoomType,
		}
		list = append(list, result.ToProto())
	}
	out := &pbroom.RoomResultListReply{
		List:  list,
		Count: rows,
	}
	return out, nil
}

func PageClubMemberRoomResultList(page *mdpage.PageOption, clubid int32, uid int32) (*pbroom.RoomResultListReply, error) {
	var list []*pbroom.RoomResults
	if page == nil {
		page.Page = 1
		page.PageSize = 20
	}
	rooms, rows, err := dbroom.PageClubMemberRoomResultList(db.DB(), uid, clubid, page)
	if err != nil {
		return nil, err
	}
	for _, mr := range rooms {
		result := &mdroom.RoomResults{
			RoomID:      mr.RoomID,
			Status:      mr.Status,
			Password:    mr.Password,
			GameType:    mr.GameType,
			CreatedAt:   mr.CreatedAt,
			RoundNow:    mr.RoundNow,
			List:        mr.UserResults,
			SubRoomType: mr.SubRoomType,
			RoomType:    mr.RoomType,
		}
		list = append(list, result.ToProto())
	}
	out := &pbroom.RoomResultListReply{
		List:  list,
		Count: rows,
	}
	return out, nil
}

func GetRoomResultByID(rid int32) (*pbroom.RoomResults, error) {
	//var result *pbroom.RoomResults
	mdr, err := dbroom.GetRoomByID(db.DB(), rid)
	if err != nil {
		return nil, err
	}
	result := &mdroom.RoomResults{
		Status:      mdr.Status,
		Password:    mdr.Password,
		GameType:    mdr.GameType,
		CreatedAt:   mdr.CreatedAt,
		RoundNow:    mdr.RoundNow,
		List:        mdr.UserResults,
		SubRoomType: mdr.SubRoomType,
		RoomType:    mdr.RoomType,
	}
	return result.ToProto(), nil
}

func CheckRoomExist(uid int32, rid int32) (int32, *mdroom.CheckRoomExist, error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		log.Err("get room user id uid:%d,rid:%d,err:%+v", uid, rid, err)
		//return 2, nil, err
	}

	//if rid == 0 || mr == nil {
	//	return 2, nil, nil
	//}

	//for _, wuid := range mr.WatchIds {
	//	if wuid == uid {
	//		mr.UserRole = enumroom.UserRoleWatch
	//		break
	//	}
	//}
	if mr == nil {
		if rid == 0 {
			err = cacheroom.DeleteAllRoomUser(mr.Password, "CheckRoomExistRoomNull")
			if err != nil {
				log.Err("room give up delete room users redis err, %v", err)
				return 4, nil, err
			}
			log.Err("check room exist ROOMNULL:%s|%d", mr.Password, uid)
			return 5, nil, nil
		} else {
			mr, err = dbroom.GetRoomByID(db.DB(), rid)
			if err != nil {
				return 8, nil, err
			}
		}
	} else if mr.Status > enumroom.RoomStatusDelay {
		err = cacheroom.DeleteAllRoomUser(mr.Password, "CheckRoomExistRoomDelay")
		if err != nil {
			log.Err("check room delete room users redis err, %v", err)
			return 6, nil, err
		}
		err = cacheroom.DeleteRoom(mr)
		if err != nil {
			log.Err("check room delete redis err, %s|%v", mr.Password, err)
			return 7, nil, err
		}
		UpdateRoom(mr)
	}
	mr.UserRole = enumroom.UserRoleWatchBro
	for _, ru := range mr.Users {
		if ru.UserID == uid {
			mr.UserRole = enumroom.UserRolePlayerBro
			break
		}
	}
	var roomStatus int32
	//fmt.Printf("AAAAAAAAARoomExistStatus:%d|%d\n", uid, mr.Status)
	inGame := false
	for _, ru := range mr.Users {
		if ru.UserID == uid {
			inGame = true
		}
	}
	if mr.Status == enumroom.RoomStatusInit {
		for _, roomuser := range mr.Users {
			if roomuser.UserID == uid {
				if roomuser.Ready == enumroom.UserUnready {
					if mr.RoundNow == 1 {
						roomStatus = enumroom.RecoveryFristInitNoReady
					} else {
						roomStatus = enumroom.RecoveryInitNoReady
					}
				} else {
					roomStatus = enumroom.RecoveryInitReady
				}
				break
			}
		}
	} else if mr.Status == enumroom.RoomStatusReInit || mr.Status == enumroom.RoomStatusAllReady {
		roomStatus = enumroom.RecoveryInitNoReady
	} else if mr.Status == enumroom.RoomStatusStarted {
		roomStatus = enumroom.RecoveryGameStart
	} else if mr.Status > enumroom.RoomStatusReInit {
		roomStatus = enumroom.RecoveryGameDone
	}
	if !inGame && mr.RoundNow > 1 && roomStatus < enumroom.RecoveryGameStart {
		roomStatus = enumroom.RecoveryWatch
	}
	Results := mdroom.RoomResults{
		RoundNumber: mr.RoundNumber,
		RoundNow:    mr.RoundNow,
		Status:      mr.Status,
		Password:    mr.Password,
		List:        mr.UserResults,
		SubRoomType: mr.SubRoomType,
		RoomType:    mr.RoomType,
	}
	//for _, ur := range mr.UserResults {
	//	fmt.Printf("AAAAACCCCC:%d|%d\n", ur.UserID, ur.Score)
	//}
	roomResults := &mdroom.CheckRoomExist{
		Result:       1,
		Room:         *mr,
		Status:       roomStatus,
		GiveupResult: mr.GiveupGame,
		GameResult:   Results,
	}
	for _, ru := range roomResults.Room.ReadyUserMap {
		roomResults.Room.Users = append(roomResults.Room.Users, ru)
	}
	return 1, roomResults, nil
}

func GetAgentRoomList(uid int32, gameType int32, page int32) (*pbroom.GetAgentRoomListReply, error) {
	var list []*pbroom.RoomResults
	f := func(r *mdroom.Room) bool {
		return true
	}
	rooms, count, _ := cacheroom.PageAgentRoom(uid, gameType, page, f)
	mpr := &mdpage.PageReply{
		PageNow:   page,
		PageTotal: count,
	}
	out := &pbroom.GetAgentRoomListReply{
		GameType:  gameType,
		PageReply: mpr.ToProto(),
	}
	if rooms == nil && len(rooms) == 0 {
		return out, nil
	}
	for _, mr := range rooms {
		result := &mdroom.RoomResults{
			RoomID:          mr.RoomID,
			Status:          mr.Status,
			Password:        mr.Password,
			GameType:        mr.GameType,
			MaxPlayerNumber: mr.MaxNumber,
			PlayerNumberNow: int32(len(mr.Users)),
			RoundNumber:     mr.RoundNumber,
			RoundNow:        mr.RoundNow,
			GameParam:       mr.GameParam,
			List:            mr.UserResults,
		}
		pbrr := result.ToProto()
		pbrr.RoomAdvanceOptions = mr.RoomAdvanceOptions
		list = append(list, pbrr)
	}
	out.List = list
	return out, nil
}

func DeleteAgentRoomRecord(uid int32, gameType int32, rid int32, pwd string) error {
	mr, err := cacheroom.GetAgentRoom(uid, gameType, rid, pwd)
	if err != nil {
		return err
	}
	if mr == nil {
		return errroom.ErrRoomNotExisted
	}

	if mr.PayerID != uid {
		return errroom.ErrNotPayer
	}
	if mr.Status < enumroom.RoomStatusDelay && mr.Status < enumroom.RoomStatusOverTimeClean {
		return errroom.ErrGameHasBegin
	}

	err = cacheroom.DeleteAgentRoom(uid, gameType, rid, pwd)
	if err != nil {
		return err
	}
	return nil
}

func DisbandAgentRoom(uid int32, pwd string) (*mdroom.Room, error) {
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errroom.ErrRoomNotExisted
	}
	if room.PayerID != uid {
		return nil, errroom.ErrNotPayer
	}
	if room.RoundNow > 1 || room.Status > enumroom.RoomStatusAllReady {
		return nil, errroom.ErrGameHasBegin
	}

	if room.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	room.Status = enumroom.RoomStatusDestroy
	room.Users = nil
	err = RoomRefund(room)
	if err != nil {
		return nil, err
	}
	err = cacheroom.DeleteAllRoomUser(room.Password, "disband agent room")
	if err != nil {
		log.Err("leave room delete all users failed, %d|%v\n",
			room.RoomID, err)
		return nil, err
	}
	err = cacheroom.DeleteRoom(room)
	if err != nil {
		log.Err("leave room delete room failed, %d|%v\n",
			room.RoomID, err)
		return nil, err
	}

	err = cacheroom.DeleteAgentRoom(room.PayerID, room.GameType, room.RoomID, room.Password)
	if err != nil {
		return nil, err
	}
	UpdateRoom(room)

	return room, nil
}

func UpdateRoom(room *mdroom.Room) error {
	f := func(tx *gorm.DB) error {
		_, err := dbroom.UpdateRoom(tx, room)
		if err != nil {
			log.Err(" update room db err, %v|%v\n", err, room)
			return err
		}
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return err
	}
	return nil
}

func GetLiveRoomCount() (int, error) {
	f := func(mdr *mdroom.Room) bool {
		if mdr.Status >= enumroom.RoomStatusInit && mdr.Status < enumroom.RoomStatusDelay {
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRooms(f)
	if rooms == nil && len(rooms) == 0 {
		return 0, nil
	}
	return len(rooms), nil
}

func ShuffleDelay() {
	rooms := cacheroom.GetAllRoomByStatus(enumroom.RoomStatusShuffle)
	if rooms == nil && len(rooms) == 0 {
		return
	}
	for _, mdr := range rooms {
		if mdr.Shuffle > 0 {
			sub := time.Now().Sub(*mdr.ShuffleAt)
			if sub.Seconds() < enumroom.ShuffleDelaySeconds {
				continue
			} else {
				mdr.Shuffle = 0
				mdr.Status = enumroom.RoomStatusAllReady
				err := cacheroom.UpdateRoom(mdr)
				if err != nil {
					log.Err("shuffle update room redis err, %v", err)
					continue
				}
			}
		}
	}
}

func AutoReady() {
	rooms := cacheroom.GetAllRoomByStatus(enumroom.RoomStatusInit)
	if rooms == nil && len(rooms) == 0 {
		return
	}
	for _, mdr := range rooms {
		//fmt.Printf("AAAAAAAAAAutoReady:%d|%d\n", mdr.GameType, mdr.RoundNow)
		if mdr.Giveup == enumroom.WaitGiveUp || mdr.GameType != enumroom.NiuniuGameType || mdr.RoundNow == 1 {
			continue
		}
		sub := time.Now().Sub(*mdr.ReadyDate)
		timeWait := float64(enumroom.ReadyWaitTime) //+ float64(len(mdr.Users))
		//fmt.Printf("AAAAAAADAutoReady:%f|%f\n", sub.Seconds(), timeWait)
		isbanker := false
		if sub.Seconds() > timeWait {
			var num int32 = 0
			for _, ru := range mdr.Users {
				if ru.UserRole != enumroom.UserRolePlayerBro {
					continue
				}

				if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
					mdclub, err := club.GetClubInfo(mdr.ClubID)
					if err != nil {
						log.Err("auto ready get club fail,uid:%d,clubid:%d,err:%v", ru.UserID, mdclub.ClubID, err)
						continue
					}
					mdcm, err := club.GetClubMember(mdr.ClubID, ru.UserID)
					if err != nil {
						log.Err("auto ready get club member fail,uid:%d,clubid:%d,err:%v", ru.UserID, mdclub.ClubID, err)
						continue
					}
					var settingParam *mdroom.SettingParam
					if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
						log.Err("room check setting param clean unmarshal room param failed, %v", err)
						continue
					}
					if mdcm.ClubCoin < settingParam.ClubCoinBaseScore || mdcm.ClubCoin < 0 {
						isbanker = checkUserIsBanker(mdr, ru)
						//fmt.Printf("CCCCCCCCCCCCC:%t|%d\n",isbanker,ru.UserID)
						if ! isbanker {
							ru.UserRole = enumroom.UserRoleSuspendBro
							//user.SuspendRound = mdr.RoundNow
							for _, rur := range mdr.UserResults {
								if rur.UserID == ru.UserID {
									rur.SuspendRound = mdr.RoundNow
								}
							}
						} else {
							//fmt.Printf("DDDDDDDDDDDD|%d\n",ru.UserID)
							continue
						}
					} else if mdcm.ClubCoin > settingParam.ClubCoinBaseScore && ru.UserRole == enumroom.UserRoleRestoreBro {
						ru.UserRole = enumroom.UserRolePlayerBro
					}
				}
				ru.Ready = enumroom.UserReady
				num++
			}
			if !isbanker && num >= enumroom.GameNumberLimit[mdr.GameType][0] {
				//fmt.Printf("EEEEEEEEEEEEEE\n")
				mdr.Status = enumroom.RoomStatusAllReady
				err := cacheroom.UpdateRoom(mdr)
				if err != nil {
					log.Err("auto ready update room redis err, %v", err)
					continue
				}
			}
			//else{
			//	t := time.Now()
			//	mdr.ReadyDate = &t
			//	err := cacheroom.UpdateRoom(mdr)
			//	if err != nil {
			//		log.Err("auto ready update room redis err, %v", err)
			//		continue
			//	}
			//}

		}
	}
}

func ReInit() []*mdroom.Room {
	rooms := cacheroom.GetAllRoomByStatus(enumroom.RoomStatusReInit)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	var mdrs []*mdroom.Room
	for _, mdr := range rooms {
		//fmt.Printf("ReInitRoomType:%d|%s\n",mdr.RoomID,mdr.RoomType)
		mdr.RoomNoticeCode = 0
		if mdr.RoomType == enumroom.RoomTypeGold {
			continue
		}
		//房间数局
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态
		if mdr.RoundNow == mdr.RoundNumber {
			if mdr.MaxNumber != mdr.StartMaxNumber {
				err := RoomRefund(mdr)
				if err != nil {
					log.Err("room done refund failed, %v", err)
				}
			}
			mdr.Status = enumroom.RoomStatusDelay
		} else {
			mdr.Status = enumroom.RoomStatusInit
			mdr.RoundNow++
			t := time.Now()
			mdr.ReadyDate = &t
			for _, user := range mdr.Users {
				user.Ready = enumroom.UserUnready
				if user.UserRole == enumroom.UserRoleWatchBro {
					user.UserRole = enumroom.UserRolePlayerBro
				} else if user.UserRole == enumroom.UserRolePlayerBro || user.UserRole == enumroom.UserRoleRestoreBro {
					if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
						mdcm, _ := club.GetClubMember(mdr.ClubID, user.UserID)
						if mdcm.Status != 1 {
							user.UserRole = enumroom.UserRoleSuspendBro
						}
						var settingParam *mdroom.SettingParam
						if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
							log.Err("room check thirteen clean unmarshal room param failed, %v", err)
							continue
						}
						if mdcm.ClubCoin < settingParam.ClubCoinBaseScore || mdcm.ClubCoin < 0 {
							if !checkUserIsBanker(mdr, user) {
								user.UserRole = enumroom.UserRoleSuspendBro
								//user.SuspendRound = mdr.RoundNow
								for _, rur := range mdr.UserResults {
									if rur.UserID == user.UserID {
										rur.SuspendRound = mdr.RoundNow - 1
									}
								}
							} else {
								mdr.RoomNoticeCode = enumroom.BankerClubCoinNoEnough
							}
						} else if mdcm.ClubCoin >= settingParam.ClubCoinBaseScore && user.UserRole == enumroom.UserRoleRestoreBro {
							user.UserRole = enumroom.UserRolePlayerBro
						}
					}
				}
			}
			for ruid, ru := range mdr.ReadyUserMap {
				//mdr.Users = append(mdr.Users, v)
				mr, err := AutoSetReady(ru, mdr) //ruid,
				if err != nil {
					log.Err("reinit sitdown fail,rid:%d,uid:%d,err:%v", mr.RoomID, ruid, err)
					cacheroom.DeleteRoomUser(ruid)
					var ids []int32
					for _, uid := range mdr.Ids {
						if uid != ruid {
							ids = append(ids, uid)
						}
					}
					mr.Ids = ids
					continue
				}
				mdr = mr
			}
			mdr.ReadyUserMap = make(map[int32]*mdroom.RoomUser)
		}

		if mdr.Status == enumroom.RoomStatusDelay {
			//游戏正常结算后 先清除玩家缓存 保留房间缓存做续费重开
			err := cacheroom.DeleteAllRoomUser(mdr.Password, "ReInitRoomDelay")
			if err != nil {
				log.Err("reinit delete all room user set redis err, %v\n", err)
				continue
			}
			if mdr.RoundNow == mdr.RoundNumber {
				err := RoundOverRoomClubCoin(mdr)
				if err != nil {
					continue
				}
			}
		}

		err := cacheroom.UpdateRoom(mdr)
		if err != nil {
			log.Err("reinit update room redis err, %v", err)
			continue
		}
		//fmt.Printf("AAAAAReInitRoom:%+v|%+v\n", time.Now(), mdr.Users)
		f := func(tx *gorm.DB) error {
			//更新玩家游戏局数
			err := dbroom.UpdateRoomPlayTimes(tx, mdr.RoomID, mdr.GameType)
			if err != nil {
				log.Err("reinit update room play times db err, %v|%v\n", err)
				return err
			}
			_, err = dbroom.UpdateRoom(tx, mdr)
			if err != nil {
				log.Err("reinit update room db err, %v|%v\n", err, mdr)
				return err
			}
			return nil
		}
		//go db.Transaction(f)
		//读写分离
		err = db.Transaction(f)
		if err != nil {
			cachelog.SetErrLog(enumroom.ServiceCode, err.Error())
			log.Err("reinit update room transaction err, %v", err)
			continue
		}
		//读写分离
		mdrs = append(mdrs, mdr)
	}
	return mdrs
}

func RoundOverRoomClubCoin(mdr *mdroom.Room) error {
	//log.Info("OOOOOOORoundOverRoomClubCoin:%d|%d|%d|%d\n", mdr.Status, mdr.RoundNow, mdr.RoomType, mdr.SubRoomType)
	if mdr.RoundNow > 1 && mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch { //mdr.Status <= enumroom.RoomStatusDelay &&

		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check setting param clean unmarshal room param failed, %v", err)
			return err
		}

		var maxScore int64 = 0
		var bigids []int32

		for _, ur := range mdr.UserResults {
			if settingParam.CostRange == enumroom.AllWinerCost {
				if ur.TotalClubCoinScore > 0 {
					bigids = append(bigids, ur.UserID)
				}
			} else if settingParam.CostRange == enumroom.BigWinerCost {
				if ur.TotalClubCoinScore > maxScore {
					maxScore = ur.TotalClubCoinScore
					bigids = []int32{}
					bigids = append(bigids, ur.UserID)
				} else if ur.TotalClubCoinScore == maxScore {
					bigids = append(bigids, ur.UserID)
				}
			}
		}
		mdr.BigWiners = bigids
		var cost float64
		if settingParam.CostType == 1 {
			tmp := fmt.Sprintf("%0.2f", float64(settingParam.CostValue)/100.0)
			cost, _ = strconv.ParseFloat(tmp, 64)
		} else {
			cost = float64(settingParam.CostValue)
		}
		costInt := int64(cost)
		//log.Info("AAAAAAAAAAAAA:%d\n", cost)
		for _, ur := range mdr.UserResults {
			//jType := tools.StringParseInt64(fmt.Sprintf("%d%d", mdr.GameType, mdr.GameIDNow))
			//jType := mdr.GameType*100 + 2
			//amount := float64(ur.TotalClubCoinScore * int64(setttingParam.ClubCoinRate))
			amount := ur.TotalClubCoinScore
			//log.Info("RoundOverRoomClubCoin:%d\n",amount)
			//log.Debug("BBBBBBGetRoomClubCoin:%d|%f|%d|%d\n", ur)
			//log.Debug("BBBBBBGetRoomClubCoin:%d|%f|%d|%d\n", ur.UserID, amount, ur.TotalClubCoinScore, setttingParam.ClubCoinRate)
			for _, uid := range mdr.BigWiners {
				if ur.UserID == uid {
					if amount < settingParam.CostBase {
						amount = 0
					} else {
						if settingParam.CostType == 1 {
							//log.Info("BBBBBBBBBBB:%d\n", amount)
							amount = int64(math.Floor(float64(amount) * cost))
							//log.Info("CCCCCCCCCCC:%d\n", amount)
							//if amount < setttingParam.CostBase {
							//	amount = 0
							//}
							//fmt.Printf("DDDDDDDDDDD:%d\n",amount)
						} else {
							//if amount > cost {
							//	amount = cost
							//}
							if amount < costInt {
								amount = 0
							} else {
								amount = costInt
							}
						}
					}

					if amount != 0 {
						mcm, err := club.GainClubMemberGameBalance(-int64(amount), mdr.ClubID, ur.UserID, int64(mdr.RoomID), int64(ur.UserID), true)
						if err != nil {
							log.Err("room club member game balance failed,rid:%d,uid:%d, err:%v", mdr.RoomID, ur.UserID, err)
							continue
						}
						for _, ru := range mdr.Users {
							if ru.UserID == ur.UserID {
								ru.ClubCoin = mcm.ClubCoin
								break
							}
						}
						//log.Debug("RoundOverRoomClubCoin:%d|%d|%+v\n", uid, amount, mcm)
					}
					break
				}
			}
		}
	}
	return nil
}

func GetRoomClubCoin(mdr *mdroom.Room) error {
	var settingParam *mdroom.SettingParam
	if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
		//log.Err("room check thirteen clean unmarshal room param failed, %v", err)
		return err
	}
	for _, ur := range mdr.UserResults {
		ru := mdr.GetRoomUser(ur.UserID)
		if ru.UserRole != enumroom.UserRolePlayerBro {
			continue
		}
		//jType := tools.StringParseInt64(fmt.Sprintf("%d%d", mdr.GameType, mdr.GameIDNow))
		//jType := mdr.GameType*100 + 2
		amount := ur.RoundScore * settingParam.ClubCoinRate //float64(ur.RoundScore * settingParam.ClubCoinRate)
		//log.Debug("AAAAAAGetRoomClubCoin:%d|%f|%d|%d\n",ur.UserID, amount, ur.RoundScore, settingParam.ClubCoinRate)
		if amount != 0 {
			mcm, err := club.GainClubMemberGameBalance(int64(amount), mdr.ClubID, ur.UserID, int64(mdr.RoomID), int64(ur.UserID), false)
			if err != nil {
				return err
			}
			ur.RoundClubCoinScore = int64(amount)
			ur.TotalClubCoinScore += ur.RoundClubCoinScore

			for _, ru := range mdr.Users {
				if ru.UserID == ur.UserID {
					ru.ClubCoin = mcm.ClubCoin
					break
				}
			}
		}
	}
	return nil
}

func GiveUpRoomDestroy() []*mdroom.Room {
	f := func(r *mdroom.Room) bool {
		if r.Giveup == enumroom.WaitGiveUp && r.Status < enumroom.RoomStatusDone {
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRooms(f)
	if len(rooms) == 0 {
		return nil
	}
	var giveRooms []*mdroom.Room
	for _, room := range rooms {
		sub := time.Now().Sub(*room.GiveupAt)
		if sub.Minutes() > enumroom.RoomGiveupCleanMinutes {
			checkRoom, err := cacheroom.GetRoom(room.Password)
			if checkRoom != nil && checkRoom.RoomID == room.RoomID {
				RoundOverRoomClubCoin(room)
				checkRoom.GiveupGame.Status = enumroom.GiveupStatusAgree
				room.Status = enumroom.RoomStatusGiveUp

				err = cacheroom.DeleteAllRoomUser(room.Password, "GiveUpRoomDestroy")
				if err != nil {
					log.Err("room give up destroy delete room users redis err, %v", err)
					continue
				}
				err = cacheroom.DeleteRoom(room)
				if err != nil {
					log.Err("room give up destroy delete room redis err, %v", err)
					continue
				}
				err = cacheroom.SetRoomDelete(room.GameType, room.RoomID)
				if err != nil {
					log.Err("give up set delete room redis err, %d|%v\n", room.RoomID, err)
				}
				f := func(tx *gorm.DB) error {
					r, err := dbroom.UpdateRoom(tx, room)
					if err != nil {
						log.Err("room give up destroy db err, %v|%v\n", err, room)
						return err
					}
					room = r
					return nil
				}
				if room.RoomType == enumroom.RoomTypeAgent {
					cacheroom.SetAgentRoom(room)
				}
				log.Debug("give up room destroy polling roomid:%d,pwd:%s,subdate:%f m\n", room.RoomID,
					room.Password, sub.Minutes())
				//go db.Transaction(f)
				//读写分离
				err = db.Transaction(f)
				if err != nil {
					cachelog.SetErrLog(enumroom.ServiceCode, err.Error())
					log.Err("room give up destroy delete room users redis err, %v", err)
					continue
				}
				//读写分离

				giveRooms = append(giveRooms, checkRoom)
			}
		}
	}
	return giveRooms
}

func DelayRoomDestroy() error {
	rooms := cacheroom.GetAllRoomByStatus(enumroom.RoomStatusDelay)
	if len(rooms) == 0 {
		return nil
	}
	for _, room := range rooms {
		sub := time.Now().Sub(*room.UpdatedAt)
		//fmt.Printf("Room Destroy Sub:%f", sub.Minutes())
		if sub.Minutes() > enumroom.RoomDelayMinutes {
			//fmt.Printf("Room Destroy Sub:%f | %d", sub.Minutes(), room.RoomID)
			checkRoom, err := cacheroom.GetRoom(room.Password)
			if checkRoom != nil && checkRoom.RoomID == room.RoomID {
				err = cacheroom.DeleteAllRoomUser(room.Password, "DelayRoomDestroy")
				if err != nil {
					log.Err("room destroy delete room users redis err, %v", err)
					continue
				}
				err = cacheroom.DeleteRoom(room)
				if err != nil {
					log.Err("room destroy delete room redis err, %v\n", err)
					continue
				}
			}
			//RoundOverRoomClubCoin(room)
			err = cacheroom.SetRoomDelete(room.GameType, room.RoomID)
			if err != nil {
				log.Err("delay set delete room redis err, %d|%v\n", room.RoomID, err)
			}
			room.Status = enumroom.RoomStatusDone
			log.Debug("DelayRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n", room.RoomID, room.Password, sub.Minutes())
			f := func(tx *gorm.DB) error {
				_, err := dbroom.UpdateRoom(tx, room)
				if err != nil {
					log.Err("room delay destroy room db err, %v\n", err)
					return err
				}
				//room = r
				return nil
			}
			if room.RoomType == enumroom.RoomTypeAgent {
				cacheroom.SetAgentRoom(room)
			}
			//go db.Transaction(f)
			err = db.Transaction(f)
			if err != nil {
				cachelog.SetErrLog(enumroom.ServiceCode, err.Error())
				log.Err("room delay room destroy delete room users redis err, %v", err)
				return err
			}
		}
	}
	return nil
}

func DeadRoomDestroy() ([]*mdroom.Room, error) {
	//定时清除不活动的房间
	f := func(r *mdroom.Room) bool {
		sub := time.Now().Sub(*r.UpdatedAt)
		//fmt.Printf("DeadRoomDestroy:%d\n",sub.Hours())
		if sub.Hours() > 24 {
			log.Debug("DeadRoomDestroyDate roomid:%d,pwd:%s,subdate:%f h\n", r.RoomID, r.Password, sub.Hours())
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRooms(f)
	if len(rooms) == 0 {
		return nil, nil
	}
	var mdrList []*mdroom.Room
	for _, room := range rooms {
		RoundOverRoomClubCoin(room)
		log.Debug("DeadRoomDestroyPolling roomid:%d,pwd:%s\n", room.RoomID, room.Password)
		err := cacheroom.DeleteAllRoomUser(room.Password, "DeadRoomDestroy")
		if err != nil {
			log.Err("delete dead room users redis err, %d|%v\n", room.RoomID, err)
			continue
		}
		err = cacheroom.DeleteRoom(room)
		if err != nil {
			log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
			continue
		}
		err = cacheroom.SetRoomDelete(room.GameType, room.RoomID)
		if err != nil {
			log.Err("delete dead set delete room redis err, %d|%v\n", room.RoomID, err)
			continue
		}
		refund := false
		if room.Status < enumroom.RoomStatusStarted {
			refund = true
		}
		room.Status = room.Status*1000 + enumroom.RoomStatusOverTimeClean
		//f := func(tx *gorm.DB) error {
		//	_, err = dbroom.UpdateRoom(tx, room)
		//	if err != nil {
		//		log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
		//	}
		//	return nil
		//}
		if room.RoomType == enumroom.RoomTypeAgent {
			cacheroom.SetAgentRoom(room)
		}

		//go db.Transaction(f)
		//err = db.Transaction(f)
		//if err != nil {
		//	log.Err("room dead room destroy delete room users redis err, %v", err)
		//	return err
		//}
		if refund {
			err = RoomRefund(room)
			if err != nil {
				log.Err("delete dead refund fail, %d|%v\n", room.RoomID, err)
			}
		}
		UpdateRoom(room)
		mdrList = append(mdrList, room)
		//err = mail.SendSysMail()

	}

	return mdrList, nil
}

func GetRoomUserLocation(user *mduser.User) ([]*pbroom.RoomUser, error) {
	mr, err := cacheroom.GetRoomUserID(user.UserID)
	if err != nil {
		return nil, err
	}
	var rus []*pbroom.RoomUser
	for _, ru := range mr.Users {
		pbru := ru.ToProto()
		userLocation := &pbroom.RoomUser{
			UserID:   ru.UserID,
			Location: pbru.Location,
		}
		rus = append(rus, userLocation)
	}
	return rus, nil
}

func checkUserIsBanker(mdr *mdroom.Room, ru *mdroom.RoomUser) bool {
	switch mdr.GameType {
	case enumroom.ThirteenGameType:
		var roomParam *mdroom.ThirteenRoomParam
		if err := json.Unmarshal([]byte(mdr.GameParam), &roomParam); err != nil {
			log.Err("check user is banker thirteen clean unmarshal room param failed, %v", err)
			return true
		}
		if roomParam.Times == 1 && ru.Role == enumroom.UserRoleMaster {
			return true
		}
		break
	case enumroom.NiuniuGameType:
		var roomParam *mdroom.NiuniuRoomParam
		if err := json.Unmarshal([]byte(mdr.GameParam), &roomParam); err != nil {
			log.Err("check user is banker unmarshal room param failed, %v", err)
			return true
		}
		if roomParam.BankerType == 4 && ru.Role == enumroom.UserRoleMaster {
			return true
		}
		if roomParam.BankerType == 2 && ru.UserID == roomParam.PreBankerID {
			return true
		}
		break
	case enumroom.FourCardGameType:
		if ru.UserID == mdr.BankerList[0] {
			return true
		}
		break
	case enumroom.TwoCardGameType:
		if ru.UserID == mdr.BankerList[0] {
			return true
		}
		break
	}
	return false
}

func checkGameParam(maxNumber int32, maxRound int32, gtype int32, gameParam string, RoomAdvanceOptions []string) (
	int32, int32, int32, string, []string, error) {
	if len(gameParam) == 0 {
		return 0, 0, 0, "", nil, errroom.ErrGameParam
	}
	if maxNumber < 2 {
		maxNumber = 2
		//return errroom.ErrRoomMaxNumber
	}
	if maxRound != 10 && maxRound != 20 && maxRound != 30 {
		maxRound = 10
		//return errroom.ErrRoomMaxRound
	}
	//fmt.Printf("ChekcGameParam:%d|%d|%d|%s\n",maxNumber,maxRound,gtype,gameParam)
	//if JoinType != enumroom.CanJoin && JoinType != enumroom.CanNotJoin {
	//	mderr := errors.Parse(errroom.ErrGameParam.Error())
	//	mderr.Detail = fmt.Sprintf(mderr.Detail, "加入类型格式错误！")
	//	return mderr
	//}

	//if len(RoomAdvanceOptions) < 3 || (RoomAdvanceOptions[0] != "1" && RoomAdvanceOptions[0] != "0" ) ||
	//	(RoomAdvanceOptions[1] != "1" && RoomAdvanceOptions[1] != "0"){
	//	RoomAdvanceOptions = []string{"0","0","1"}
	//}
	lenrao := len(RoomAdvanceOptions)
	if lenrao > 3 {
		RoomAdvanceOptions = RoomAdvanceOptions[:3]
	}
	for i := 0; i < 3; i++ {
		if lenrao < i+1 {
			if i == 2 {
				RoomAdvanceOptions = append(RoomAdvanceOptions, "1")
			} else {
				RoomAdvanceOptions = append(RoomAdvanceOptions, "0")
			}
		} else {
			if i < 2 && RoomAdvanceOptions[i] != "0" && RoomAdvanceOptions[i] != "1" {
				RoomAdvanceOptions[i] = "0"
			} else if i == 2 && RoomAdvanceOptions[i] != "1" && RoomAdvanceOptions[i] != "2" &&
				RoomAdvanceOptions[i] != "5" && RoomAdvanceOptions[i] != "10" && RoomAdvanceOptions[i] != "0.1" &&
				RoomAdvanceOptions[i] != "0.2" && RoomAdvanceOptions[i] != "0.5" {
				RoomAdvanceOptions[i] = "1"
			}
		}
	}

	switch gtype {
	case enumroom.ThirteenGameType:
		if maxNumber > 8 {
			maxNumber = 8
			//return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.ThirteenRoomParam
		mderr := errors.Parse(errroom.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		//if roomParam.BankerType != 1 && roomParam.BankerType != 2 {
		//	return errors.ErrGameParam
		//}
		if roomParam.BankerAddScore < 0 || roomParam.BankerAddScore > 6 || roomParam.BankerAddScore%2 != 0 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "当庄加分格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Joke != 0 && roomParam.Joke != 1 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "大小王格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Times < 1 || roomParam.Times > 3 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "比赛模式格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.NiuniuGameType:
		if maxNumber != 4 && maxNumber != 6 && maxNumber != 8 && maxNumber != 10 {
			return 0, 0, 0, "", nil, errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.NiuniuRoomParam
		mderr := errors.Parse(errroom.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BankerType < 1 || roomParam.BankerType > 5 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "玩法ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Times != 1 && roomParam.Times != 2 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "倍数ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BetScore < 1 || roomParam.BetScore > 4 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "底分ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		//if len(roomParam.SpecialCards) == 0 {
		//	//mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型长度错误！")
		//	//return 0, 0, 0, "", nil, mderr
		//	roomParam.SpecialCards = []string{"0", "0", "0", "0", "0", "0", "0"}
		//}
		lenSC := len(roomParam.SpecialCards)
		if lenSC > 7 {
			roomParam.SpecialCards = roomParam.SpecialCards[:7]
		}
		for i := 0; i < 7; i++ {
			if lenSC < i+1 {
				roomParam.SpecialCards = append(roomParam.SpecialCards, "0")
			} else {
				if roomParam.SpecialCards[i] != "0" && roomParam.SpecialCards[i] != "1" {
					roomParam.SpecialCards[i] = "0"
				}
			}
		}

		lenAo := len(roomParam.AdvanceOptions)
		if lenAo > 3 {
			roomParam.AdvanceOptions = roomParam.AdvanceOptions[:3]
		}
		for i := 0; i < 3; i++ {
			if lenAo < i+1 {
				roomParam.AdvanceOptions = append(roomParam.AdvanceOptions, "0")
			} else {
				if i == 0 && roomParam.AdvanceOptions[i] != "0" && roomParam.AdvanceOptions[i] != "1" &&
					roomParam.AdvanceOptions[i] != "2" && roomParam.AdvanceOptions[i] != "3" {
					roomParam.AdvanceOptions[i] = "0"
				} else if i > 0 && roomParam.AdvanceOptions[i] != "0" && roomParam.AdvanceOptions[i] != "1" {
					roomParam.AdvanceOptions[i] = "0"
				}
			}
		}

		//if len(roomParam.AdvanceOptions) != 3 {
		//	mderr.Detail = fmt.Sprintf(mderr.Detail, "高级选项长度错误！")
		//	return 0, 0, 0, "", nil, mderr
		//}
		//
		//for _, value := range roomParam.SpecialCards {
		//	if value != "1" && value != "0" {
		//		mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型格式错误！")
		//		return 0, 0, 0, "", nil, mderr
		//	}
		//}

		if roomParam.AdvanceOptions[0] != "0" && roomParam.BankerType == 5 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择推注和通比！")
			return 0, 0, 0, "", nil, mderr
		}

		//if roomParam.AdvanceOptions[0] != "0" && roomParam.AdvanceOptions[0] != "1" && roomParam.AdvanceOptions[0] != "2" && roomParam.AdvanceOptions[0] != "3" {
		//	mderr.Detail = fmt.Sprintf(mderr.Detail, "推注最高倍数格式错误！")
		//	return 0, 0, 0, "", nil, mderr
		//}

		if roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1" {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和不发花牌！")
			return 0, 0, 0, "", nil, mderr
		}

		if maxNumber == 10 && roomParam.AdvanceOptions[1] == "1" { //|| (roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1")
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和10人模式！")
			return 0, 0, 0, "", nil, mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)

		break
	case enumroom.DoudizhuGameType:
		if maxNumber != 4 {
			maxNumber = 4
			//return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.DoudizhuRoomParam
		mderr := errors.Parse(errroom.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("doudizhu unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BaseScore != 0 && roomParam.BaseScore != 5 && roomParam.BaseScore != 10 {
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "基本分格式错误！")
			//return mderr
			roomParam.BaseScore = 0
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.FourCardGameType:
		if maxNumber > 8 {
			maxNumber = 8
			//return errroom.ErrRoomMaxNumber
		}
		mderr := errors.Parse(errroom.ErrGameParam.Error())
		var roomParam *mdroom.FourCardRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("fourcard unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, errroom.ErrGameParam
		}
		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
			roomParam.ScoreType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
			//return mderr

		}
		if roomParam.BetType < 1 || roomParam.BetType > 2 {
			roomParam.BetType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
			//return mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.TwoCardGameType:
		if maxNumber > 10 {
			maxNumber = 10
		}
		var roomParam *mdroom.TwoCardRoomParam
		mderr := errors.Parse(errroom.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("towcard unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
			roomParam.ScoreType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
			//return errroom.ErrGameParam
		}
		if roomParam.BetType < 1 || roomParam.BetType > 2 {
			roomParam.BetType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
			//return mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.RunCardGameType:

		break
	default:
		return 0, 0, 0, "", nil, errroom.ErrGameParam
	}

	return maxNumber, maxRound, gtype, gameParam, RoomAdvanceOptions, nil
}

func CheckHasRoom(uid int32) (bool, *mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(uid)
	if hasRoom {
		mdr, err := cacheroom.GetRoomUserID(uid)
		if err != nil {
			mderr := errors.Parse(err.Error())
			if mderr.Code == 40001 {
				err = cacheroom.DeleteRoomUser(uid)
				if err != nil {
					log.Err("room check room exist delete room users redis err, %v", err)
					return false, nil, err
				}
				log.Err("check room exist delete room users redis fail uid:%d,err:%v", uid, err)
				return false, nil, nil
			}
			return hasRoom, nil, err
		}
		return hasRoom, mdr, err
	}
	return false, nil, nil
}

func GameStart(uid int32) (*mdroom.Room, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}
	if uid != mdr.Users[0].UserID {
		return nil, errroom.ErrNotPayer
	}
	if mdr.Status > enumroom.RoomStatusInit {
		return nil, errroom.ErrGameHasBegin
	}
	if len(mdr.Users) < 2 {
		return nil, errroom.ErrPlayerNumberNoEnough
	}
	allReady := true
	for _, user := range mdr.Users {
		if allReady && user.Ready != enumroom.UserReady {
			allReady = false
		}
	}
	if allReady {
		if mdr.RoundNow == 1 {
			mdr.Users[0].Role = enumroom.UserRoleMaster
		}
		mdr.Status = enumroom.RoomStatusAllReady
	} else {
		return nil, errroom.ErrRoomNotAllReady
	}
	//_, mdPayer := cacheuser.GetUserByID(mdr.PayerID)
	//mdr.StartMaxNumber = mdr.MaxNumber
	mdr.MaxNumber = int32(len(mdr.Users))
	//cost := getRoomCost(mdr.GameType, mdr.StartMaxNumber, mdr.RoundNumber, mdPayer.Channel, mdPayer.Version, mdPayer.MobileOs, mdr.RoomType)
	//mdr.Cost = cost
	err = RoomCreateBalance(mdr, mdr.PayerID)
	if err != nil {
		log.Err("RoomCreateBalanceERR:%+v", err)
		return nil, err
	}
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
		return nil, err
	}
	return mdr, nil
}

func RoomCreateBalance(mdr *mdroom.Room, uid int32) error {
	//jType := getRoomJournalType(mdr.GameType)

	if mdr.Cost != 0 {
		if mdr.RoomType == enumroom.RoomTypeClub {
			//if mdr.SubRoomType == enumroom.SubTypeClubMatch {
			//	return nil
			//}
			err := club.SetClubGameBalance(-mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, mdr.GameType*100+1,
				int64(mdr.RoomID), int64(uid))
			if err != nil {
				return err
			}
		} else {
			if mdr.RoomAdvanceOptions[1] == "0" {
				err := bill.GainGameBalance(mdr.PayerID, mdr.RoomID, mdr.GameType*100+1, &mbill.Balance{Amount:
				-mdr.Cost, CoinType: enumcom.Diamond})
				if err != nil {
					log.Err("back room balance err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
						mdr.PayerID, mdr.Cost, mdr.CostType, err)
					return err
				}
			} else {

				avgCost := mdr.Cost / int64(len(mdr.Users))
				var ids []int32
				var passIds []int32
				for _, ru := range mdr.Users {
					mdb, err := bill.GetUserBalance(ru.UserID, enumcom.Diamond)
					if err != nil {
						return err
					}
					//fmt.Printf("BBBBRoomCreateBalance:%d|%f|%f\n", ru.UserID, mdb.Amount, avgCost)
					if mdb.Amount < avgCost {
						passIds = append(passIds, ru.UserID)
					} else {
						ids = append(ids, ru.UserID)
					}
				}
				if len(passIds) > 0 {
					mderr := errors.Parse(errroom.ErrUserDiamondNoEnough.Error())
					b := &bytes.Buffer{}
					for _, uid := range passIds {
						_,mdu:= cacheuser.GetUserByID(uid)
						b.WriteString(mdu.Nickname)
						b.WriteString(",")
					}
					mderr.Detail = fmt.Sprintf(mderr.Detail, b.String(), avgCost)
					return mderr
				}
				for _, uid := range ids {
					err := bill.GainGameBalance(uid, mdr.RoomID, mdr.GameType*100+1, &mbill.Balance{Amount:
					-avgCost, CoinType: enumcom.Diamond})

					if err != nil {
						log.Err("back room balance err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
							mdr.PayerID, avgCost, mdr.CostType, err)
						return err
					}
				}
			}
		}
	}
	return nil
}

func RoomRefund(mdr *mdroom.Room) error {

	//if mdr.Cost > 0 && ((mdr.RoundNow == 1 && mdr.Status > enumroom.RoomStatusAllReady) || (mdr.RoundNow > 1)) {
	//	jType := mdr.GameType*100 + 2 //getRoomJournalType(mdr.GameType)
	//	//_, mdPayer := cacheuser.GetUserByID(mdr.PayerID)
	//	//cost := getRoomCost(mdr.GameType, mdr.MaxNumber, mdr.RoundNumber, mdPayer.Channel, mdPayer.Version, mdPayer.MobileOs, mdr.RoomType)
	//	amount := mdr.Cost //- cost
	//
	//	if amount == 0 {
	//		return nil
	//	}
	//	f := func(tx *gorm.DB) error {
	//		if mdr.RoomType == enumroom.RoomTypeClub {
	//
	//			err := club.SetClubGameBalance(amount, enumbill.TypeDiamond, mdr.ClubID, jType, int64(mdr.RoomID),
	//				int64(mdr.PayerID))
	//			if err != nil {
	//				return err
	//			}
	//		} else {
	//			err := bill.GainGameBalance(mdr.PayerID, mdr.RoomID, jType, &mbill.Balance{Amount:
	//			amount, CoinType: enumcom.Diamond})
	//			if err != nil {
	//				log.Err("back room balance err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
	//					mdr.PayerID, amount, mdr.CostType, err)
	//				return err
	//			}
	//		}
	//		return nil
	//	}
	//	err := db.Transaction(f)
	//	if err != nil {
	//		cachelog.SetErrLog(enumroom.ServiceCode, err.Error())
	//		return err
	//	}
	//}
	return nil
}

//func RoomCreateBalance(mdr *mdroom.Room, mdu *mduser.User) error {
//	//jType := getRoomJournalType(mdr.GameType)
//	if mdr.Cost != 0 {
//		if mdr.RoomType == enumroom.RoomTypeClub {
//			//if mdr.SubRoomType == enumroom.SubTypeClubMatch {
//			//	return nil
//			//}
//			err := club.SetClubGameBalance(-mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, mdr.GameType*100+1,
//				int64(mdr.RoomID), int64(mdu.UserID))
//			if err != nil {
//				return err
//			}
//		} else {
//			_, err := bill.SetBalanceFreeze(mdu.UserID, int64(mdr.RoomID), &mbill.Balance{Amount: mdr.Cost,
//				CoinType: enumcom.Diamond}, mdr.GameType*100+2)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func roomBackUnFreezeAndBalance(mdr *mdroom.Room) error {
//	//jType := getRoomJournalType(mdr.GameType)
//	if mdr.Cost != 0 {
//		if mdr.RoomType != enumroom.RoomTypeClub {
//			//if mdr.SubRoomType == enumroom.SubTypeClubMatch {
//			//	return nil
//			//}
//			_, err := bill.SetBalanceFreeze(mdr.PayerID, int64(mdr.RoomID),
//				&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond}, mdr.GameType*100+2)
//			if err != nil {
//				log.Err("back room freeze err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
//					mdr.PayerID, -mdr.Cost, mdr.CostType, err)
//				return err
//			}
//			err = bill.GainGameBalance(mdr.PayerID, mdr.RoomID, mdr.GameType*100+1, &mbill.Balance{Amount:
//			-mdr.Cost, CoinType: enumcom.Diamond})
//			if err != nil {
//				log.Err("back room balance err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
//					mdr.PayerID, mdr.Cost, mdr.CostType, err)
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func roomStartBalance(mdr *mdroom.Room, mdu *mduser.User) error {
//	//jType := getRoomJournalType(mdr.GameType)
//	if mdr.Cost != 0 {
//		if mdr.RoomType == enumroom.RoomTypeClub {
//			//if mdr.SubRoomType == enumroom.SubTypeClubMatch {
//			//	return nil
//			//}
//			err := club.SetClubGameBalance(-mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, mdr.GameType*100+1,
//				int64(mdr.RoomID), int64(mdu.UserID))
//			if err != nil {
//				return err
//			}
//		} else {
//			err := bill.GainGameBalance(mdu.UserID, mdr.RoomID, mdr.GameType*100+1,
//				&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond})
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func RoomRefund(mdr *mdroom.Room) error {
//	if mdr.Cost != 0 {
//		jType := mdr.GameType*100 + 2 //getRoomJournalType(mdr.GameType)
//		f := func(tx *gorm.DB) error {
//			if mdr.RoomType == enumroom.RoomTypeClub {
//				//if mdr.SubRoomType == enumroom.SubTypeClubMatch {
//				//	return nil
//				//}
//				err := club.SetClubGameBalance(mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, jType, int64(mdr.RoomID),
//					int64(mdr.PayerID))
//				if err != nil {
//					return err
//				}
//			} else {
//				_, err := bill.SetBalanceFreeze(mdr.PayerID, int64(mdr.RoomID),
//					&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond}, jType)
//				if err != nil {
//					log.Err("back room cost err roomid:%d,payerid:%d,cost:%d,constype:%d", mdr.RoomID,
//						mdr.PayerID, mdr.Cost, mdr.CostType)
//					return err
//				}
//			}
//			return nil
//		}
//		err := db.Transaction(f)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func PageSpecialGameList(page *mdpage.PageOption, plgr *mdroom.PlayerSpecialGameRecord) (
	[]*mdroom.PlayerSpecialGameRecord, int64, error) {
	return dbroom.PageSpecialGameList(db.DB(), plgr, page)
}

func SetBankerList(pwd string, mduser *mduser.User) (*mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(mduser.UserID)
	if !hasRoom {
		return nil, errroom.ErrUserNotInRoom
	}
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	//if mdr.ClubID > 0 && mduser.ClubID != mdr.ClubID {
	//	return nil, errroom.ErrNotClubMember
	//}
	if mdr.Status > enumroom.RoomStatusDelay {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}

	for _, wuid := range mdr.WatchIds {
		if wuid == mduser.UserID {
			return nil, errroom.ErrNotInGame
		}
	}
	for uid, _ := range mdr.ReadyUserMap {
		if uid == mduser.UserID {
			return nil, errroom.ErrInReadyMap
		}
	}
	for _, uid := range mdr.BankerList {
		if uid == mduser.UserID {
			return nil, errroom.ErrAlreadyInBankerList
		}
	}
	if mdr.RoomType == enumroom.RoomTypeGold {
		limit := enumroom.GoldRoomCostMap[mdr.GameType][mdr.Level][1]
		userBalance, err := bill.GetUserBalance(mduser.UserID, enumbill.TypeGold)
		if err != nil {
			return nil, err
		}
		if userBalance.Balance < limit {
			return nil, errroom.ErrNotEnoughGold
		}
	} else if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, err := club.GetClubMember(mdr.ClubID, mduser.UserID)
		if err != nil {
			return nil, err
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return nil, mderr
		}
	}
	mdr.BankerList = append(mdr.BankerList, mduser.UserID)

	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("banker list join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mduser.UserID)
	if err != nil {
		log.Err("banker list join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	//UpdateRoom(mdr)
	return mdr, nil
}

func OutBankerList(pwd string, mduser *mduser.User) (*mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(mduser.UserID)
	if !hasRoom {
		return nil, errroom.ErrUserNotInRoom
	}
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	//if mdr.ClubID > 0 && mduser.ClubID != mdr.ClubID {
	//	return nil, errroom.ErrNotClubMember
	//}
	if mdr.Status > enumroom.RoomStatusDelay {
		return nil, errroom.ErrGameIsDone
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return nil, errroom.ErrInGiveUp
	}
	//if mdr.BankerList == nil || len(mdr.BankerList)== 0{
	//	return nil, errroom.ErrNoInBankerList
	//}
	//if mdr.BankerList[0] == mduser.UserID && (mdr.Status != enumroom.RoomStatusInit &&
	//	mdr.Status > enumroom.RoomStatusReInit ) {
	//	return nil, errroom.ErrOutBankerListWithBanker
	//}
	if len(mdr.BankerList) == 1 && mdr.BankerList[0] == mduser.UserID {
		return nil, errroom.ErrOutBankerList
	}

	var ids []int32
	inList := false
	for _, uid := range mdr.BankerList {
		if uid != mduser.UserID {
			ids = append(ids, uid)
		} else {
			inList = true
		}
	}
	if !inList {
		return nil, errroom.ErrNoInBankerList
	}
	mdr.BankerList = ids
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("out banker list join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mduser.UserID)
	if err != nil {
		log.Err("out banker join set session failed, %v|%v\n", err, mdr)
		return nil, err
	}
	//UpdateRoom(mdr)
	return mdr, nil
}

func GetVipRoomList(clubid int32, sid int32) ([]*mdroom.Room, error) {
	return cacheroom.GetAllVipRoom(clubid, sid)
}

func GetRoomRoundNow(gtype int32) ([]*mdroom.ClubRoomLog, error) {
	return dbroom.GetRoomRoundNow(db.DB(), gtype)
}

func SetSuspend(uid int32) (*mdroom.Room, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, err
	}
	if mdr.Status > enumroom.RoomStatusInit {
		return nil, errroom.ErrGameHasBegin
	}
	ru := mdr.GetRoomUser(uid)
	if ru == nil {
		return nil, errroom.ErrUserNotInRoom
	}
	if ru.UserRole != enumroom.UserRolePlayerBro {
		return nil, errroom.ErrNotInGame
	}
	ru.UserRole = enumroom.UserRoleSuspendBro
	//mdr.WatchIds = append(mdr.WatchIds, uid)
	err = UpdateRoom(mdr)
	if err != nil {
		return nil, err
	}
	return mdr, err
}

func SetRestore(uid int32) (*mdroom.Room, *mdroom.RoomUser, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, nil, err
	}

	ru := mdr.GetRoomUser(uid)
	if ru == nil {
		return nil, nil, errroom.ErrUserNotInRoom
	}
	if ru.UserRole != enumroom.UserRoleSuspendBro {
		return nil, nil, errroom.ErrNotInSuspend
	}
	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, uid)
		if mdcm.Status != 1 {
			return nil, nil, errroom.ErrCanNotIntoClubRoom
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return nil, nil, errroom.ErrGameParam
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			mderr := errors.Parse(errroom.ErrNotEnoughClubCoin.Error())
			mderr.Detail = fmt.Sprintf(mderr.Detail, settingParam.ClubCoinBaseScore)
			return nil, nil, mderr
		}
		ru.ClubCoin = mdcm.ClubCoin
	}
	if mdr.Status > enumroom.RoomStatusAllReady && mdr.Status < enumroom.RoomStatusReInit {
		ru.UserRole = enumroom.UserRoleRestoreBro
	} else if mdr.Status == enumroom.RoomStatusInit || mdr.Status == enumroom.RoomStatusAllReady {
		ru.UserRole = enumroom.UserRolePlayerBro
	} else if mdr.Status > enumroom.RoomStatusReInit {
		return nil, nil, errroom.ErrGameIsDone
	}
	//sub := time.Now().Sub(*mdr.ReadyDate)
	//if sub.Seconds() < 2 {
	//	t1 := mdr.ReadyDate
	//	s, _ := time.ParseDuration("2s")
	//	t2 := t1.Add(s)
	//	mdr.ReadyDate = &t2
	//}

	t1 := mdr.ReadyDate
	s, _ := time.ParseDuration("2s")
	t2 := t1.Add(s)
	mdr.ReadyDate = &t2
	//mdr.WatchIds = append(mdr.WatchIds, uid)
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		return nil, nil, err
	}
	return mdr, ru, err
}

func CheckUserClubCoin(mdr *mdroom.Room, uid int32) bool {
	if mdr.RoomType == enumroom.RoomTypeClub && mdr.SubRoomType == enumroom.SubTypeClubMatch {
		mdcm, _ := club.GetClubMember(mdr.ClubID, uid)
		if mdcm.Status != 1 {
			return false
		}
		var settingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mdr.SettingParam), &settingParam); err != nil {
			log.Err("room check user club coin unmarshal room param failed, %v", err)
			return false
		}
		if mdcm.ClubCoin < settingParam.ClubCoinBaseScore {
			return false
		}
	}
	return true
}

func GetRoomListByIDS(uds []int32) []*pbroom.Room {
	var pbrs []*pbroom.Room
	for _, uid := range uds {
		mdr, err := cacheroom.GetRoomUserID(uid)
		if err != nil {
			continue
		}
		pbr := &pbroom.Room{
			Password: mdr.Password,
			RoomID:   mdr.RoomID,
			GameType: mdr.GameType,
			UserID:   uid,
		}
		pbrs = append(pbrs, pbr)
	}
	return pbrs
}
