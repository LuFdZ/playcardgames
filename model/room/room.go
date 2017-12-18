package room

import (
	"math/rand"
	"playcards/model/bill"
	"encoding/json"
	enumb "playcards/model/bill/enum"
	mbill "playcards/model/bill/mod"
	"playcards/model/club"
	mdpage "playcards/model/page"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	"playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	cacheu "playcards/model/user/cache"
	mdu "playcards/model/user/mod"
	pbr "playcards/proto/room"
	enumcom "playcards/model/common/enum"
	"playcards/model/config"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"fmt"
)

func GenerateRangeNum(min, max int) string {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	return strconv.Itoa(randNum)
}

func GetRoomByUserID(uid int32) (*mdr.Room, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return nil, errors.ErrUserNotInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if mr == nil {
		return nil, errors.ErrRoomNotExisted
	}
	return mr, nil
}

func RenewalRoom(pwd string, user *mdu.User) (int32, []int32, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return 0, nil, nil, errors.ErrUserAlreadyInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, nil, nil, err
	}
	if mr == nil {
		return 0, nil, nil, errors.ErrRoomNotExisted
	}
	if mr.ClubID > 0 {
		return 0, nil, nil, errors.ErrClubCantRenewal
	}
	if mr.RoomType == enumr.RoomTypeAgent {
		return 0, nil, nil, errors.ErrRoomType
	}
	if mr.Status != enumr.RoomStatusDelay {
		return 0, nil, nil, errors.ErrRenewalRoon
	}
	var ids []int32
	for _, id := range mr.Ids {
		if id != user.UserID {
			ids = append(ids, id)
		}
	}
	oldID := mr.RoomID
	newRoom, err := CreateRoom(mr.RoomType, mr.GameType, mr.MaxNumber,
		mr.RoundNumber, mr.GameParam, user, pwd)
	if err != nil {
		log.Err("room Renewal create new room failed,%v | %v\n", mr, err)
		return 0, nil, nil, err
	}
	cacher.DeleteRoom(mr.Password)
	err = cacher.UpdateRoom(newRoom)
	if err != nil {
		log.Err("room Renewal set redis failed,%v | %v\n", mr, err)
		return 0, nil, nil, err
	}
	return oldID, ids, newRoom, nil
}

func CreateRoom(rtype int32, gtype int32, maxNum int32, roundNum int32,
	gParam string, user *mdu.User, pwd string) (*mdr.Room,
	error) {
	clubID := user.ClubID
	var err error
	checkPwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkPwd) != 0 {
		return nil, errors.ErrUserAlreadyInRoom
	}
	if rtype == enumr.RoomTypeClub && clubID == 0 {
		return nil, errors.ErrNotClubMember
	}

	err = ChekcGameParam(maxNum,roundNum,gtype,gParam)
	if err != nil {
		return nil,err
	}
	users := []*mdr.RoomUser{}
	ids := []int32{}
	if rtype == 0 {
		rtype = enumr.RoomTypeNom
	}
	if rtype == enumr.RoomTypeNom || rtype == enumr.RoomTypeClub {
		roomUser := GetRoomUser(user, enumr.UserUnready, 1,
			enumr.UserRoleMaster)
		users = append(users, roomUser)
		ids = append(ids, user.UserID)
	} else {
		f := func(r *mdr.Room) bool {
			if r.Status < enumr.RoomStatusStarted {
				return true
			}
			return false
		}
		_, _, total := cacher.PageAgentRoom(user.UserID, enumr.AgentRoomAllGameType, enumr.AgentRoomAllPage, f)
		if total >= enumr.AgentRoomLimit {
			return nil, errors.ErrRoomAgentLimit
		}
	}
	cost := getRoomCost(gtype, maxNum, roundNum, user, rtype)

	if len(pwd) == 0 {
		pwd, err = getPassWord()
		if err != nil {
			return nil, err
		}
	}
	now := gorm.NowFunc()
	//cost := -diamond
	//if gtype == enumr.DoudizhuGameType{
	//	maxNum = 4
	//}
	mr := &mdr.Room{
		Password:    pwd,
		GameType:    gtype,
		MaxNumber:   maxNum,
		RoundNumber: roundNum,
		RoundNow:    1,
		GameParam:   gParam,
		Status:      enumr.RoomStatusInit,
		Giveup:      enumr.NoGiveUp,
		Users:       users,
		RoomType:    rtype,
		PayerID:     user.UserID,
		GiveupAt:    &now,
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Ids:         ids,
		Cost:        cost,
		CostType:    enumr.CostTypeDiamond,
		Flag:        enumr.RoomNoFlag,
	}
	if mr.RoomType == enumr.RoomTypeClub {
		mr.ClubID = clubID
	}
	jType := getRoomJournalType(mr.GameType)
	f := func(tx *gorm.DB) error {
		err := dbr.CreateRoom(tx, mr)
		if err != nil {
			log.Err("room create failed,%v | %v", mr, err)
			return err
		}

		if cost != 0 {
			if rtype == enumr.RoomTypeClub {
				err = club.SetClubBalance(-cost, enumb.TypeDiamond, clubID, jType, int64(mr.RoomID), int64(user.UserID))
				if err != nil {
					return err
				}
			} else {
				_, err = bill.SetBalanceFreeze(user.UserID, int64(mr.RoomID), &mbill.Balance{Amount: cost, CoinType: enumcom.Diamond}, jType)
				if err != nil {
					return err
				}
			}
		}

		err = cacher.SetRoom(mr)
		if err != nil {
			log.Err("room create set redis failed,%v | %v\n", mr, err)
			return err
		}
		if rtype == enumr.RoomTypeNom || rtype == enumr.RoomTypeClub {
			err = cacher.SetRoomUser(mr.RoomID, mr.Password, user.UserID)
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

func getPassWord() (string, error) {
	pwdNew := GenerateRangeNum(enumr.RoomCodeMin, enumr.RoomCodeMax)
	exist := cacher.CheckRoomExist(pwdNew)
	for i := 0; exist && i < 3; i++ {
		pwdNew = GenerateRangeNum(enumr.RoomCodeMin, enumr.RoomCodeMax)
		exist = cacher.CheckRoomExist(pwdNew)
	}
	if exist {
		return "", errors.ErrRoomPwdExisted
	}
	return pwdNew, nil
}

func getRoomCost(gType int32, maxNumber int32, roundNumber int32, user *mdu.User, roomtype int32) int64 {
	var diamond int64
	var cost int32
	if gType == enumr.ThirteenGameType {
		cost = enumr.ThirteenGameCost
	} else if gType == enumr.NiuniuGameType {
		cost = enumr.NiuniuGameCost
	} else if gType == enumr.DoudizhuGameCost {
		cost = enumr.DoudizhuGameCost
	}

	diamond = int64(maxNumber * roundNumber * cost)
	if roomtype == enumr.RoomTypeNom {
		rate := config.CheckConfigCondition(user.Channel, user.Version, user.MobileOs)
		diamond = int64(rate * float64(diamond))
	}
	return diamond
}

func getRoomJournalType(gameType int32) int32 {
	var jType int32
	if gameType == enumr.ThirteenGameType {
		jType = enumb.JournalTypeThirteenFreeze
	} else if gameType == enumr.NiuniuGameType {
		jType = enumb.JournalTypeNiuniuFreeze
	} else if gameType == enumr.DoudizhuGameType {
		jType = enumb.JournalTypeDoudizhuFreeze
	}
	return jType
}

func JoinRoom(pwd string, user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, nil, errors.ErrUserAlreadyInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if mr == nil {
		return nil, nil, errors.ErrRoomNotExisted
	}
	if mr.ClubID > 0 && user.ClubID != mr.ClubID {
		return nil, nil, errors.ErrNotClubMember
	}
	if mr.Status != enumr.RoomStatusInit {
		return nil, nil, errors.ErrNotReadyStatus
	}
	if mr.Giveup == enumr.WaitGiveUp {
		return nil, nil, errors.ErrInGiveUp
	}

	num := len(mr.Users)
	if num >= (int)(mr.MaxNumber) {
		return nil, nil, errors.ErrRoomFull
	}
	p := 0
	//var positionArray [room.MaxNumber]int32
	positionArray := make([]int32, mr.MaxNumber)
	//positionArray :=[room.MaxNumber]int32{}
	for n := 0; n < len(positionArray); n++ {
		positionArray[n] = 0
	}
	for _, ru := range mr.Users {
		if ru.UserID == user.UserID {
			return nil, nil, errors.ErrUserAlreadyInRoom
		}
		positionArray[ru.Position-1] = 1
	}
	for n := 0; n < len(positionArray); n++ {
		if positionArray[n] == 0 {
			p = n
		}
	}
	//role := enumr.UserRoleSlave
	//if mr.RoomType == enumr.RoomTypeAgent && num == 0 {
	//	role = enumr.UserRoleMaster
	//}
	roomUser := GetRoomUser(user, enumr.UserUnready, int32(p+1),
		enumr.UserRoleSlave)
	newUsers := append(mr.Users, roomUser)
	mr.Users = newUsers
	mr.Ids = append(mr.Ids, user.UserID)

	err = cacher.UpdateRoom(mr)
	if err != nil {
		log.Err("room jooin set session failed, %v|%v\n", err, mr)
		return nil, nil, err
	}
	err = cacher.SetRoomUser(mr.RoomID, mr.Password, user.UserID)
	if err != nil {
		log.Err("room user jooin set session failed, %v|%v\n", err, mr)
		return nil, nil, err
	}

	//balance, err := mbill.GetUserBalance(user.UserID)
	//if err != nil {
	//	return nil, nil, err
	//}
	//rate := mbill.CheckConfigCondition(user.Channel, user.Version, user.MobileOs)
	//needDiamond := int64(rate * float64(int64(mr.MaxNumber*mr.RoundNumber*enumr.ThirteenGameCost)))
	//if balance.Diamond < needDiamond {
	//	mr.Status = enumr.RoomStatusDiamondNoEnough
	//	err = cacher.DeleteAllRoomUser(mr.Password, "JoinRoomDiamondNoEnough")
	//	if err != nil {
	//		log.Err("join room diamond no enough delete all users failed, %d|%v\n",
	//			mr.RoomID, err)
	//		return nil, nil, err
	//	}
	//	err = cacher.DeleteRoom(mr.Password)
	//	if err != nil {
	//		log.Err("join room diamond no enough room failed, %d|%v\n",
	//			mr.RoomID, err)
	//		return nil, nil, err
	//	}
	//
	//}
	UpdateRoom(mr)
	return roomUser, mr, nil
}

func LeaveRoom(user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	//pwd := cacher.GetRoomPasswordByUserID(user.UserID)
	//if len(pwd) == 0 {
	//	return nil, nil, errors.ErrUserNotInRoom
	//}
	//room, err := cacher.GetRoom(pwd)
	mr, err := GetRoomByUserID(user.UserID)
	if err != nil {
		return nil, nil, err
	}

	if mr.RoundNumber > 1 && mr.Status > enumr.RoomStatusAllReady {
		return nil, nil, errors.ErrGameHasBegin
	}

	if mr.Giveup == enumr.WaitGiveUp {
		return nil, nil, errors.ErrInGiveUp
	}

	newUsers := []*mdr.RoomUser{}
	roomUser := &mdr.RoomUser{}
	handle := 0
	masterLeave := false
	for _, u := range mr.Users {
		if u.UserID == user.UserID {
			handle = 1
			roomUser = u
			if u.Role == enumr.UserRoleMaster &&
				(mr.RoomType == enumr.RoomTypeNom) {
				log.Info("delete room cause master leave.user:%d,room:%d\n",
					user.UserID, mr.RoomID)
				masterLeave = true
			}
		} else {
			newUsers = append(newUsers, u)
		}
	}
	mr.Users = newUsers

	var ids []int32
	for _, user := range mr.Users {
		ids = append(ids, user.UserID)
	}
	mr.Ids = ids
	if handle == 0 {
		return nil, nil, errors.ErrUserNotInRoom
	}
	//RoomTypeNom 普通开房解散条件 人员全部退出 || 房主退出
	//RoomTypeAgent 代开房解散条件 代开房者主动解散
	//RoomTypeClub 俱乐部开房解散条件 人员全部退出
	if (len(newUsers) == 0 && mr.RoomType != enumr.RoomTypeAgent) ||
		(masterLeave && mr.RoomType == enumr.RoomTypeNom) {
		mr.Users = nil
		mr.Status = enumr.RoomStatusDestroy
		err = cacher.DeleteAllRoomUser(mr.Password, "LeaveRoom")
		if err != nil {
			log.Err("leave room delete all users failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacher.DeleteRoom(mr.Password)
		if err != nil {
			log.Err("leave room delete room failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		RoomRefund(mr)

	} else {
		err := cacher.UpdateRoom(mr)
		if err != nil {
			log.Err("room leave room delete failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		err = cacher.DeleteRoomUser(user.UserID)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				mr.RoomID, err)
			return nil, nil, err
		}
		UpdateRoom(mr)
	}
	return roomUser, mr, nil
}

func RoomRefund(mr *mdr.Room) {
	if mr.Cost != 0 {
		var jType int32
		if mr.GameType == enumr.ThirteenGameType {
			jType = enumb.JournalTypeThirteenUnFreeze
		} else if mr.GameType == enumr.NiuniuGameType {
			jType = enumb.JournalTypeNiuniuUnFreeze
		}else if mr.GameType == enumr.DoudizhuGameType {
			jType = enumb.JournalTypeDoudizhuUnFreeze
		}
		f := func(tx *gorm.DB) error {
			if mr.RoomType == enumr.RoomTypeClub {
				err := club.SetClubBalance(mr.Cost, enumb.TypeDiamond, mr.ClubID, jType, int64(mr.RoomID), int64(mr.PayerID))
				if err != nil {
					return err
				}
			} else {
				_, err := bill.SetBalanceFreeze(mr.PayerID, int64(mr.RoomID),
					&mbill.Balance{Amount: -mr.Cost, CoinType: enumcom.Diamond}, jType)
				if err != nil {
					log.Err("BackRoomCostErr roomid:%d,payerid:%d,cost:%d,constype:%d", mr.RoomID, mr.PayerID, mr.Cost, mr.CostType)
					return err
				}
			}
			return nil
		}
		go db.Transaction(f)
	}
}

func GetReady(pwd string, uid int32) (*mdr.RoomUser, []int32, error) {
	checkPwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkPwd) == 0 || len(pwd) == 0 || pwd != checkPwd {
		return nil, nil, errors.ErrUserNotInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if mr.Status > enumr.RoomStatusInit {
		return nil, nil, errors.ErrNotReadyStatus
	}
	if mr.Giveup == enumr.WaitGiveUp {
		return nil, nil, errors.ErrInGiveUp
	}
	allReady := true
	t := time.Now()
	var num int32
	for _, user := range mr.Users {
		num++
		if user.UserID == uid {
			if user.Ready == enumr.UserReady {
				return nil, nil, errors.ErrAlreadyReady
			}
			user.Ready = enumr.UserReady
			user.UpdatedAt = &t
		}
		if allReady && user.Ready != enumr.UserReady {
			allReady = false
		}
	}
	if allReady && num == mr.MaxNumber {
		if mr.RoundNow == 1 {
			mr.Users[0].Role = enumr.UserRoleMaster
		}
		mr.Status = enumr.RoomStatusAllReady
	}
	err = cacher.UpdateRoom(mr)
	if err != nil {
		log.Err("room ready failed, roomid:%d,uid:%d,err:%v\n", mr.RoomID, uid, err)
		return nil, nil, err
	}

	readyUser := &mdr.RoomUser{
		UserID: uid,
	}

	return readyUser, mr.Ids, nil
}

func GiveUpGame(pwd string, uid int32) ([]int32, *mdr.GiveUpGameResult, *mdr.Room,
	error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, nil, nil, errors.ErrUserNotInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, nil, err
	}

	if mr.RoundNumber == 1 && mr.Status < enumr.RoomStatusStarted {
		return nil, nil, nil, errors.ErrNotReadyStatus
	}

	if mr.Giveup == enumr.WaitGiveUp {
		return GiveUpVote(pwd, 1, uid)
	}

	giveUpResult := mr.GiveupGame
	var list []*mdr.UserState
	agreeGiveUp := 0
	for _, user := range mr.Users {
		var state int32
		scoketstatus := cacher.GetUserStatus(user.UserID)
		if user.UserID == uid {
			state = enumr.UserStateLeader
			agreeGiveUp++
		} else if scoketstatus == enumr.SocketClose {
			state = enumr.UserStateOffline
			//agreeGiveUp++
		} else {
			state = enumr.UserStateWaiting
		}
		us := &mdr.UserState{
			UserID: user.UserID,
			State:  state,
		}
		list = append(list, us)
	}
	if agreeGiveUp == len(mr.Users) {
		mr.Status = enumr.RoomStatusGiveUp
		giveUpResult.Status = enumr.GiveupStatusAgree
	} else {
		mr.Giveup = enumr.WaitGiveUp
		now := gorm.NowFunc()
		mr.GiveupAt = &now
		giveUpResult.Status = enumr.GiveupStatusWairting
	}

	giveUpResult.UserStateList = list
	mr.GiveupGame = giveUpResult
	if mr.Status == enumr.RoomStatusGiveUp {
		err = cacher.DeleteAllRoomUser(mr.Password, "GiveUpGame")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil, nil, nil, err
		}
		err = cacher.DeleteRoom(mr.Password)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
		err = cacher.SetRoomDelete(mr.GameType, mr.RoomID)
		if err != nil {
			log.Err("give up set delete room redis err, %d|%v\n", mr.RoomID, err)
		}
		if mr.RoomType == enumr.RoomTypeAgent {
			cacher.SetAgentRoom(mr)
		}
		UpdateRoom(mr)
		//RoomRefund(mr)
	} else {
		err = cacher.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	}

	log.Info("give up game:%d|%v", uid, giveUpResult)
	return mr.Ids, &giveUpResult, mr, nil
}

func GiveUpVote(pwd string, status int32, uid int32) ([]int32, *mdr.GiveUpGameResult, *mdr.Room,
	error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, nil, nil, errors.ErrUserNotInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, nil, err
	}

	if mr.Giveup != enumr.WaitGiveUp {
		return nil, nil, nil, errors.ErrNotInGiveUp
	}

	giveUpResult := mr.GiveupGame
	if status == 1 {
		status = enumr.UserStateAgree
	} else {
		status = enumr.UserStateDisagree
	}
	giveup := enumr.GiveupStatusWairting
	agreeGiveUp := 0
	for _, userstate := range mr.GiveupGame.UserStateList {
		if userstate.UserID == uid {
			if userstate.State != enumr.UserStateWaiting && userstate.State != enumr.UserStateOffline {
				return nil, nil, nil, errors.ErrAlreadyVoted
			}
			userstate.State = status
			if userstate.State == enumr.UserStateDisagree {
				giveup = enumr.GiveupStatusDisAgree
				break
			}

		} else if userstate.State == enumr.UserStateDisagree {
			giveup = enumr.GiveupStatusDisAgree
			break
		}

		if userstate.State != enumr.UserStateDisagree &&
			userstate.State != enumr.UserStateWaiting {
			agreeGiveUp++
		}
	}
	if agreeGiveUp == len(mr.GiveupGame.UserStateList) {
		mr.Status = enumr.RoomStatusGiveUp
		giveUpResult.Status = enumr.GiveupStatusAgree
		err = cacher.DeleteAllRoomUser(mr.Password, "GiveUpVote")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil, nil, nil, err
		}
		err = cacher.DeleteRoom(mr.Password)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
		err = cacher.SetRoomDelete(mr.GameType, mr.RoomID)
		if err != nil {
			log.Err("give up set delete room redis err, %d|%v\n", mr.RoomID, err)
		}
		if mr.RoomType == enumr.RoomTypeAgent {
			cacher.SetAgentRoom(mr)
		}
		UpdateRoom(mr)
		//RoomRefund(mr)
	} else if giveup == enumr.GiveupStatusDisAgree {
		mr.Giveup = enumr.NoGiveUp
		giveUpResult.Status = enumr.GiveupStatusDisAgree
		err = cacher.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	} else {
		mr.Giveup = enumr.WaitGiveUp
		giveUpResult.Status = enumr.GiveupStatusWairting
		err = cacher.UpdateRoom(mr)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, nil, nil, err
		}
	}

	log.Info("give up game vote:%d|%v", uid, giveUpResult)
	return mr.Ids, &giveUpResult, mr, nil
}

func GetRoomUser(u *mdu.User, ready int32, position int32,
	role int32) *mdr.RoomUser {
	return &mdr.RoomUser{
		UserID: u.UserID,
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
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return errors.ErrUserNotInRoom
	}
	pwdCheck := cacher.GetRoomPasswordByUserID(sendid)
	if len(pwd) == 0 {
		return errors.ErrUserNotInRoom
	}

	if pwd != pwdCheck {
		return errors.ErrNotInSameRoon
	}

	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return err
	}
	if mr.Status != enumr.RoomStatusStarted {
		return errors.ErrNotReadyStatus
	}

	return nil
}

func VoiceChat(uid int32) (*mdr.Room, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return nil, errors.ErrUserNotInRoom
	}
	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if mr.Status > enumr.RoomStatusDone {
		return nil, errors.ErrGameIsDone
	}
	return mr, nil
}

func TestClean() error {
	cacher.FlushAll()
	f := func(tx *gorm.DB) error {
		dbr.DeleteAll(tx)
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return err
	}
	return nil
}

func PageFeedbackList(page *mdpage.PageOption, fb *mdr.Feedback) (
	[]*mdr.Feedback, int64, error) {
	return dbr.PageFeedbackList(db.DB(), page, fb)
}

func CreateFeedback(fb *mdr.Feedback, uid int32, ip string) (*mdr.Feedback, error) {
	fb.UserID = uid
	fb.LoginIP = ip
	return dbr.CreateFeedback(db.DB(), fb)
}

func RoomResultList(page *mdpage.PageOption, uid int32, gtype int32) (*pbr.RoomResultListReply, error) {
	var list []*pbr.RoomResults
	if page == nil {
		page.Page = 1
		page.PageSize = 20
	}
	rooms, rows, err := dbr.PageRoomResultList(db.DB(), uid, gtype, page)
	if err != nil {
		return nil, err
	}
	for _, mr := range rooms {
		result := &mdr.RoomResults{
			Status:    mr.Status,
			Password:  mr.Password,
			GameType:  mr.GameType,
			CreatedAt: mr.CreatedAt,
			RoundNow:  mr.RoundNow,
			List:      mr.UserResults,
		}
		list = append(list, result.ToProto())
	}
	out := &pbr.RoomResultListReply{
		List:  list,
		Count: rows,
	}
	return out, nil
}

func CheckRoomExist(uid int32) (int32, *mdr.CheckRoomExist, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		log.Err("check room exist pwd null:%s|%d", pwd, uid)
		return 2, nil, nil
	}

	mr, err := cacher.GetRoom(pwd)
	if err != nil {
		return 3, nil, err
	}
	if mr == nil {
		err = cacher.DeleteAllRoomUser(pwd, "CheckRoomExistRoomNull")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return 4, nil, err
		}
		log.Err("CheckRoomExistROOMNULL:%s|%d", pwd, uid)
		return 5, nil, nil
	} else if mr.Status > enumr.RoomStatusDelay {
		err = cacher.DeleteAllRoomUser(mr.Password, "CheckRoomExistRoomDelay")
		if err != nil {
			log.Err("check room delete room users redis err, %v", err)
			return 6, nil, err
		}
		err = cacher.DeleteRoom(pwd)
		if err != nil {
			log.Err("check room delete redis err, %s|%v", pwd, err)
			return 7, nil, err
		}
		UpdateRoom(mr)
	}

	var roomStatus int32
	if mr.Status == enumr.RoomStatusInit {
		for _, roomuser := range mr.Users {
			if roomuser.UserID == uid {
				if roomuser.Ready == enumr.UserUnready {
					if mr.RoundNow == 1 {
						roomStatus = enumr.RecoveryFristInitNoReady
					} else {
						roomStatus = enumr.RecoveryInitNoReady
					}
				} else {
					roomStatus = enumr.RecoveryInitReady
				}
				break
			}
		}
	} else if mr.Status == enumr.RoomStatusReInit {
		roomStatus = enumr.RecoveryInitNoReady
	} else {
		roomStatus = enumr.RecoveryGameStart
	}
	Results := mdr.RoomResults{
		RoundNumber: mr.RoundNumber,
		RoundNow:    mr.RoundNow,
		Status:      mr.Status,
		Password:    mr.Password,
		List:        mr.UserResults,
	}
	roomResults := &mdr.CheckRoomExist{
		Result:       1,
		Room:         *mr,
		Status:       roomStatus,
		GiveupResult: mr.GiveupGame,
		GameResult:   Results,
	}

	return 1, roomResults, nil
}

func RoomUserStatusCheck() []*pbr.UserConnection {
	var ucs []*pbr.UserConnection
	f := func(r *mdr.Room) bool {
		if r.Status == enumr.RoomStatusStarted {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	for _, mr := range rooms {
		for _, user := range mr.Users {
			//status := cacher.GetUserStatus(user.UserID)
			statusNow := cacheu.GetUserOnlineStatus(user.UserID)
			if statusNow == 0 {
				statusNow = enumr.SocketClose
			} else {
				statusNow = enumr.SocketAline
			}
			statusOld := cacher.GetUserStatus(user.UserID)
			//若房间不在游戏开始状态 或者连接状态未初始化 或者已广播过 不做处理
			if (mr.Status != enumr.RoomStatusInit &&
				mr.Status != enumr.RoomStatusStarted &&
				mr.Status != enumr.RoomStatusAllReady &&
				mr.Status != enumr.RoomStatusReInit) ||
				statusNow == statusOld {
				continue
			}
			uc := &pbr.UserConnection{
				Ids:    mr.Ids,
				UserID: user.UserID,
				Status: statusNow,
			}
			cacher.UpdateRoomUserSocektStatus(user.UserID, statusNow)
			ucs = append(ucs, uc)
		}
	}
	return ucs
}

func GetAgentRoomList(uid int32, gameType int32, page int32) (*pbr.GetAgentRoomListReply, error) {
	var list []*pbr.RoomResults
	f := func(r *mdr.Room) bool {
		return true
	}
	rooms, count, _ := cacher.PageAgentRoom(uid, gameType, page, f)
	mpr := &mdpage.PageReply{
		PageNow:   page,
		PageTotal: count,
	}
	out := &pbr.GetAgentRoomListReply{
		GameType:  gameType,
		PageReply: mpr.ToProto(),
	}
	if rooms == nil && len(rooms) == 0 {
		return out, nil
	}
	for _, mr := range rooms {
		result := &mdr.RoomResults{
			RoomID:          mr.RoomID,
			Status:          mr.Status,
			Password:        mr.Password,
			GameType:        mr.GameType,
			CreatedAt:       mr.CreatedAt,
			MaxPlayerNumber: mr.MaxNumber,
			PlayerNumberNow: int32(len(mr.Users)),
			RoundNumber:     mr.RoundNumber,
			RoundNow:        mr.RoundNow,
			GameParam:       mr.GameParam,
			List:            mr.UserResults,
		}

		list = append(list, result.ToProto())
	}
	out.List = list
	return out, nil
}

func DeleteAgentRoomRecord(uid int32, gameType int32, rid int32, pwd string) error {
	mr, err := cacher.GetAgentRoom(uid, gameType, rid, pwd)
	if err != nil {
		return err
	}
	if mr == nil {
		return errors.ErrRoomNotFind
	}

	if mr.PayerID != uid {
		return errors.ErrNotPayer
	}
	if mr.Status < enumr.RoomStatusDelay {
		return errors.ErrGameHasBegin
	}

	err = cacher.DeleteAgentRoom(uid, gameType, rid, pwd)
	if err != nil {
		return err
	}
	return nil
}

func DisbandAgentRoom(uid int32, pwd string) (*mdr.Room, error) {
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrRoomNotExisted
	}
	if room.PayerID != uid {
		return nil, errors.ErrNotPayer
	}
	if room.RoundNumber > 1 && room.Status > enumr.RoomStatusAllReady {
		return nil, errors.ErrGameHasBegin
	}

	if room.Giveup == enumr.WaitGiveUp {
		return nil, errors.ErrInGiveUp
	}
	room.Status = enumr.RoomStatusDestroy
	room.Users = nil
	err = cacher.DeleteAllRoomUser(room.Password, "disband agent room")
	if err != nil {
		log.Err("leave room delete all users failed, %d|%v\n",
			room.RoomID, err)
		return nil, err
	}
	err = cacher.DeleteRoom(room.Password)
	if err != nil {
		log.Err("leave room delete room failed, %d|%v\n",
			room.RoomID, err)
		return nil, err
	}

	err = cacher.DeleteAgentRoom(room.PayerID, room.GameType, room.RoomID, room.Password)
	if err != nil {
		return nil, err
	}
	UpdateRoom(room)
	RoomRefund(room)
	return room, nil
}

func UpdateRoom(room *mdr.Room) error {
	f := func(tx *gorm.DB) error {
		_, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			log.Err(" update room db err, %v|%v\n", err, room)
			return err
		}
		return nil
	}
	go db.Transaction(f)
	return nil
}

func GetLiveRoomCount() (int, error) {
	f := func(r *mdr.Room) bool {
		if r.Status >= enumr.RoomStatusInit && r.Status < enumr.RoomStatusDelay {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if rooms == nil && len(rooms) == 0 {
		return 0, nil
	}
	return len(rooms), nil
}

func ReInit() []*mdr.Room {

	f := func(r *mdr.Room) bool {
		if r.Status == enumr.RoomStatusReInit {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	for _, room := range rooms {
		//房间数局
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态
		if room.RoundNow == room.RoundNumber {
			room.Status = enumr.RoomStatusDelay
		} else {
			room.Status = enumr.RoomStatusInit
			room.RoundNow++
			for _, user := range room.Users {
				user.Ready = enumr.UserUnready
			}
		}

		if room.Status == enumr.RoomStatusDelay {
			//游戏正常结算后 先清除玩家缓存 保留房间缓存做续费重开
			err := cacher.DeleteAllRoomUser(room.Password, "ReInitRoomDelay")
			if err != nil {
				log.Err("reinit delete all room user set redis err, %v\n", err)
			}
		}
		err := cacher.SetRoom(room)
		if err != nil {
			log.Err("reinit update room redis err, %v", err)
			continue
		}

		f := func(tx *gorm.DB) error {
			//更新玩家游戏局数
			err := dbr.UpdateRoomPlayTimes(tx, room.RoomID, room.GameType)
			if err != nil {
				log.Err("reinit update room play times db err, %v|%v\n", err)
				return err
			}
			_, err = dbr.UpdateRoom(tx, room)
			if err != nil {
				log.Err("reinit update room db err, %v|%v\n", err, room)
				return err
			}
			return nil
		}
		//go db.Transaction(f)
		//读写分离
		err = db.Transaction(f)
		if err != nil {
			log.Err("reinit update room transaction err, %v", err)
			continue
		}
		//读写分离
	}
	return rooms
}

func GiveUpRoomDestroy() []*mdr.Room {

	f := func(r *mdr.Room) bool {
		if r.Giveup == enumr.WaitGiveUp && r.Status < enumr.RoomStatusDone {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if len(rooms) == 0 {
		return nil
	}
	var giveRooms []*mdr.Room
	for _, room := range rooms {
		sub := time.Now().Sub(*room.GiveupAt)
		if sub.Minutes() > enumr.RoomGiveupCleanMinutes {
			checkRoom, err := cacher.GetRoom(room.Password)
			if checkRoom != nil && checkRoom.RoomID == room.RoomID {
				checkRoom.GiveupGame.Status = enumr.GiveupStatusAgree
				room.Status = enumr.RoomStatusGiveUp
				err = cacher.DeleteAllRoomUser(room.Password, "GiveUpRoomDestroy")
				if err != nil {
					log.Err("room give up destroy delete room users redis err, %v", err)
					continue
				}
				err = cacher.DeleteRoom(room.Password)
				if err != nil {
					log.Err("room give up destroy delete room redis err, %v", err)
					continue
				}
				err = cacher.SetRoomDelete(room.GameType, room.RoomID)
				if err != nil {
					log.Err("give up set delete room redis err, %d|%v\n", room.RoomID, err)
				}
				f := func(tx *gorm.DB) error {
					r, err := dbr.UpdateRoom(tx, room)
					if err != nil {
						log.Err("room give up destroy db err, %v|%v\n", err, room)
						return err
					}
					room = r
					return nil
				}
				if room.RoomType == enumr.RoomTypeAgent {
					cacher.SetAgentRoom(room)
				}
				log.Debug("GiveUpRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n", room.RoomID, room.Password, sub.Minutes())
				//go db.Transaction(f)
				//读写分离
				err = db.Transaction(f)
				if err != nil {
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
	f := func(r *mdr.Room) bool {
		if r.Status == enumr.RoomStatusDelay {
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if len(rooms) == 0 {
		return nil
	}
	for _, room := range rooms {
		sub := time.Now().Sub(*room.UpdatedAt)
		//fmt.Printf("Room Destroy Sub:%f", sub.Minutes())
		if sub.Minutes() > enumr.RoomDelayMinutes {
			//fmt.Printf("Room Destroy Sub:%f | %d", sub.Minutes(), room.RoomID)
			checkRoom, err := cacher.GetRoom(room.Password)
			if checkRoom != nil && checkRoom.RoomID == room.RoomID {
				err = cacher.DeleteAllRoomUser(room.Password, "DelayRoomDestroy")
				if err != nil {
					log.Err("room destroy delete room users redis err, %v", err)
					continue
				}
				err = cacher.DeleteRoom(room.Password)
				if err != nil {
					log.Err("room destroy delete room redis err, %v\n", err)
					continue
				}
			}
			err = cacher.SetRoomDelete(room.GameType, room.RoomID)
			if err != nil {
				log.Err("delay set delete room redis err, %d|%v\n", room.RoomID, err)
			}
			room.Status = enumr.RoomStatusDone
			log.Debug("DelayRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n", room.RoomID, room.Password, sub.Minutes())
			f := func(tx *gorm.DB) error {
				_, err := dbr.UpdateRoom(tx, room)
				if err != nil {
					log.Err("room delay destroy room db err, %v\n", err)
					return err
				}
				//room = r
				return nil
			}
			if room.RoomType == enumr.RoomTypeAgent {
				cacher.SetAgentRoom(room)
			}
			//go db.Transaction(f)
			err = db.Transaction(f)
			if err != nil {
				log.Err("room delay room destroy delete room users redis err, %v", err)
				return err
			}
		}
	}
	return nil
}

func DeadRoomDestroy() error {
	//定时清除不活动的房间
	f := func(r *mdr.Room) bool {
		sub := time.Now().Sub(*r.UpdatedAt)
		//fmt.Printf("DeadRoomDestroy:%d\n",sub.Hours())
		if sub.Hours() > 24 {
			log.Debug("DeadRoomDestroyDate roomid:%d,pwd:%s,subdate:%f h\n", r.RoomID, r.Password, sub.Hours())
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if len(rooms) == 0 {
		return nil
	}
	for _, room := range rooms {
		log.Debug("DeadRoomDestroyPolling roomid:%d,pwd:%s\n", room.RoomID, room.Password)
		err := cacher.DeleteAllRoomUser(room.Password, "DeadRoomDestroy")
		if err != nil {
			log.Err("delete dead room users redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacher.DeleteRoom(room.Password)
		if err != nil {
			log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacher.SetRoomDelete(room.GameType, room.RoomID)
		if err != nil {
			log.Err("delete dead set delete room redis err, %d|%v\n", room.RoomID, err)
		}
		room.Status = room.Status*10 + enumr.RoomStatusOverTimeClean
		f := func(tx *gorm.DB) error {
			_, err = dbr.UpdateRoom(tx, room)
			if err != nil {
				log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
			}
			return nil
		}
		if room.RoomType == enumr.RoomTypeAgent {
			cacher.SetAgentRoom(room)
		}

		//go db.Transaction(f)
		err = db.Transaction(f)
		if err != nil {
			log.Err("room dead room destroy delete room users redis err, %v", err)
			return err
		}
		if room.Status < enumr.RoomStatusStarted {
			RoomRefund(room)
		}

	}

	return nil
}

func GetRoomUserLocation(user *mdu.User) ([]*pbr.RoomUser, error) {
	mr, err := GetRoomByUserID(user.UserID)
	if err != nil {
		return nil, err
	}
	var rus []*pbr.RoomUser
	for _, ru := range mr.Users {
		pbru := ru.ToProto()
		userLocation := &pbr.RoomUser{
			UserID:   ru.UserID,
			Location: pbru.Location,
		}
		rus = append(rus, userLocation)
	}
	return rus, nil
}

func ChekcGameParam(maxNumber int32, maxRound int32, gtype int32, gameParam string) error {
	if len(gameParam) == 0{
		return errors.ErrGameParam
	}
	if maxNumber < 2{
		return errors.ErrRoomMaxNumber
	}
	if maxRound != 10 && maxRound != 20 && maxRound != 30 {
		return errors.ErrRoomMaxRound
	}
	//fmt.Printf("ChekcGameParam:%d|%d|%d|%s\n",maxNumber,maxRound,gtype,gameParam)
	switch gtype {
	case enumr.ThirteenGameType:
		if maxNumber > 4 {
			return errors.ErrRoomMaxNumber
		}
		var roomParam *mdr.ThirteenRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return errors.ErrGameParam
		}
		//if roomParam.BankerType != 1 && roomParam.BankerType != 2 {
		//	return errors.ErrGameParam
		//}
		if roomParam.BankerAddScore < 0 || roomParam.BankerAddScore > 6 || roomParam.BankerAddScore%2 != 0 {
			return errors.ErrGameParam
		}
		if roomParam.Joke != 0 && roomParam.Joke != 1 {
			return errors.ErrGameParam
		}
		if roomParam.Times <1 || roomParam.Times >3 {
			return errors.ErrGameParam
		}
		break
	case enumr.NiuniuGameType:
		if maxNumber > 5 {
			return errors.ErrRoomMaxNumber
		}
		var roomParam *mdr.NiuniuRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			return errors.ErrGameParam
		}
		if roomParam.BankerType < 1 || roomParam.BankerType > 4 {
			return errors.ErrGameParam
		}
		if roomParam.Times != 3 && roomParam.Times != 5 && roomParam.Times != 10 {
			return errors.ErrGameParam
		}
		break
	case enumr.DoudizhuGameType:
		if maxNumber != 4 {
			return errors.ErrRoomMaxNumber
		}
		var roomParam *mdr.DoudizhuRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("doudizhu unmarshal room param failed, %v", err)
			return errors.ErrGameParam
		}
		if roomParam.BaseScore != 0 && roomParam.BaseScore != 5 && roomParam.BaseScore != 10 {
			fmt.Printf("DoudizhuGameType BaseScore:%v\n",roomParam)
			return errors.ErrGameParam
		}

	default:
		return errors.ErrGameParam
	}
	return nil
}
