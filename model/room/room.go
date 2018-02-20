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
	//"playcards/model/mail"
	"playcards/model/config"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
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
	newRoom, err := CreateRoom(mr.RoomType, mr.GameType, mr.StartMaxNumber,
		mr.RoundNumber, mr.GameParam, mduser, pwd)
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

func CreateRoom(rtype int32, gtype int32, maxNum int32, roundNum int32,
	gParam string, user *mduser.User, pwd string) (*mdroom.Room,
	error) {
	clubID := user.ClubID
	var err error
	hasRoom := cacheroom.ExistRoomUser(user.UserID)
	if hasRoom {
		return nil, errroom.ErrUserAlreadyInRoom
	}
	if rtype == enumroom.RoomTypeClub && clubID == 0 {
		return nil, errroom.ErrNotClubMember
	}

	if cacheroom.GetRoomTestConfigKey("CheckGameParam") != "0" {
		err = chekcGameParam(maxNum, roundNum, gtype, gParam)
		if err != nil {
			return nil, err
		}
	}

	users := []*mdroom.RoomUser{}
	ids := []int32{}
	if rtype == 0 {
		rtype = enumroom.RoomTypeNom
	}
	channel := user.Channel
	version := user.Version
	mobileOs := user.MobileOs
	payerID := user.UserID
	if rtype == enumroom.RoomTypeClub {
		mdclub, _ := club.GetClubFromDB(clubID)
		_, creater := cacheuser.GetUserByID(mdclub.CreatorID)
		if creater == nil {
			return nil, errroom.ErrClubCreaterNotFind
		}
		channel = creater.Channel
		version = creater.Version
		mobileOs = creater.MobileOs
		payerID = mdclub.CreatorID
	}
	cost := getRoomCost(gtype, maxNum, roundNum, channel, version, mobileOs, rtype)

	if rtype == enumroom.RoomTypeNom {
		roomUser := GetRoomUser(user, enumroom.UserUnready, 1,
			enumroom.UserRoleMaster)
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
		roomUser := GetRoomUser(user, enumroom.UserUnready, 1,
			enumroom.UserRoleMaster)
		users = append(users, roomUser)
		ids = append(ids, user.UserID)
		mdclub, err := club.GetClubFromDB(clubID)
		if err != nil {
			return nil, err
		}
		if mdclub.Diamond < cost {
			return nil, errroom.ErrNotEnoughDiamond
		}
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
		Password:       pwd,
		GameType:       gtype,
		MaxNumber:      maxNum,
		RoundNumber:    roundNum,
		RoundNow:       1,
		GameParam:      gParam,
		Status:         enumroom.RoomStatusInit,
		Giveup:         enumroom.NoGiveUp,
		Users:          users,
		RoomType:       rtype,
		PayerID:        payerID,
		GiveupAt:       &now,
		Ids:            ids,
		Cost:           cost,
		StartMaxNumber: maxNum,
		CostType:       enumroom.CostTypeDiamond,
		Flag:           enumroom.RoomNoFlag,
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

		err = RoomCreateBalance(mr, user)
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
	var cost int32
	var configID int32
	if roomtype == enumroom.RoomTypeNom {
		configID = enumroom.ConsumeOpen
	} else if roomtype == enumroom.RoomTypeAgent {
		configID = enumroom.AgentConsumeOpen
	} else if roomtype == enumroom.RoomTypeClub {
		configID = enumroom.ClubConsumeOpen
	}

	if gType == enumroom.ThirteenGameType {
		cost = enumroom.ThirteenGameCost
	} else if gType == enumroom.NiuniuGameType {
		cost = enumroom.NiuniuGameCost
	} else if gType == enumroom.DoudizhuGameType {
		cost = enumroom.DoudizhuGameCost
	} else if gType == enumroom.FourCardGameType {
		cost = enumroom.FourcardGameCost
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

func JoinRoom(pwd string, mduser *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(mduser.UserID)
	if hasRoom {
		return nil, nil, errroom.ErrUserAlreadyInRoom
	}
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if mdr.ClubID > 0 && mduser.ClubID != mdr.ClubID {
		return nil, nil, errroom.ErrNotClubMember
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
	roomUser := GetRoomUser(mduser, enumroom.UserUnready, int32(p+1),
		enumroom.UserRoleSlave)
	if mdr.RoomType != enumroom.RoomTypeNom && len(mdr.Users) == 0 {
		roomUser.Role = enumroom.UserRoleMaster
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

func LeaveRoom(mduser *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	mr, err := cacheroom.GetRoomUserID(mduser.UserID)
	if err != nil {
		return nil, nil, err
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

	var ids []int32
	for _, user := range mr.Users {
		ids = append(ids, user.UserID)
	}
	mr.Ids = ids
	if handle == 0 {
		return nil, nil, errroom.ErrUserNotInRoom
	}
	if roomUser.Role == enumroom.UserRoleMaster && mr.RoomType != enumroom.RoomTypeNom && len(mr.Users) >0 {
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

		RoomRefund(mr)
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

func GetReady(pwd string, uid int32, shuffle bool) (bool, *mdroom.Room, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return false, nil, err
	}
	if pwd != mdr.Password {
		return false, nil, errroom.ErrUserNotInRoom
	}
	if mdr.Status > enumroom.RoomStatusInit {
		return false, nil, errroom.ErrNotReadyStatus
	}
	if mdr.Giveup == enumroom.WaitGiveUp {
		return false, nil, errroom.ErrInGiveUp
	}
	if shuffle {
		//if !isBeginOrEndRound(mdr, uid) {
		//	return false,nil, errroom.ErrShuffle
		//}
		mdr.Shuffle = uid
	}
	allReady := true
	t := time.Now()
	var num int32
	for _, user := range mdr.Users {
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
	if allReady && num == mdr.MaxNumber {
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
	if allReady && num == mdr.MaxNumber && mdr.RoundNow == 1 {
		err = roomBackUnFreezeAndBalance(mdr)
		if err != nil {
			log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
			return false, nil, err
		}
	}
	return allReady, mdr, nil
}

func GiveUpGame(pwd string, uid int32) ([]int32, *mdroom.GiveUpGameResult, *mdroom.Room,
	error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, nil, nil, err
	}

	if mr.RoundNumber == 1 && mr.Status < enumroom.RoomStatusStarted {
		return nil, nil, nil, errroom.ErrNotReadyStatus
	}

	if mr.Giveup == enumroom.WaitGiveUp {
		return GiveUpVote(pwd, 1, uid)
	}

	giveUpResult := mr.GiveupGame
	var list []*mdroom.UserState
	agreeGiveUp := 0
	for _, user := range mr.Users {
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
		UpdateRoom(mr)
		//RoomRefund(mr)
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

func GetRoomUser(mdu *mduser.User, ready int32, position int32,
	role int32) *mdroom.RoomUser {
	return &mdroom.RoomUser{
		UserID: mdu.UserID,
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
			Status:    mr.Status,
			Password:  mr.Password,
			GameType:  mr.GameType,
			CreatedAt: mr.CreatedAt,
			RoundNow:  mr.RoundNow,
			List:      mr.UserResults,
		}
		list = append(list, result.ToProto())
	}
	out := &pbroom.RoomResultListReply{
		List:  list,
		Count: rows,
	}
	return out, nil
}

func CheckRoomExist(uid int32, rid int32) (int32, *mdroom.CheckRoomExist, error) {
	mr, err := cacheroom.GetRoomUserID(uid)
	if err != nil && rid == 0 {
		return 2, nil, err
	}

	if mr == nil {
		if rid == 0 {
			err = cacheroom.DeleteAllRoomUser(mr.Password, "CheckRoomExistRoomNull")
			if err != nil {
				log.Err("room give up delete room users redis err, %v", err)
				return 4, nil, err
			}
			log.Err("CheckRoomExistROOMNULL:%s|%d", mr.Password, uid)
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

	var roomStatus int32
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
	Results := mdroom.RoomResults{
		RoundNumber: mr.RoundNumber,
		RoundNow:    mr.RoundNow,
		Status:      mr.Status,
		Password:    mr.Password,
		List:        mr.UserResults,
	}
	roomResults := &mdroom.CheckRoomExist{
		Result:       1,
		Room:         *mr,
		Status:       roomStatus,
		GiveupResult: mr.GiveupGame,
		GameResult:   Results,
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
	if mr.Status < enumroom.RoomStatusDelay &&  mr.Status < enumroom.RoomStatusOverTimeClean{
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
	RoomRefund(room)
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

func ReInit() []*mdroom.Room {
	rooms := cacheroom.GetAllRoomByStatus(enumroom.RoomStatusReInit)
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	for _, mdr := range rooms {
		//房间数局
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态

		if mdr.RoundNow == mdr.RoundNumber {
			mdr.Status = enumroom.RoomStatusDelay
		} else {
			mdr.Status = enumroom.RoomStatusInit
			mdr.RoundNow++
			for _, user := range mdr.Users {
				user.Ready = enumroom.UserUnready
			}
		}

		if mdr.Status == enumroom.RoomStatusDelay {
			//游戏正常结算后 先清除玩家缓存 保留房间缓存做续费重开
			err := cacheroom.DeleteAllRoomUser(mdr.Password, "ReInitRoomDelay")
			if err != nil {
				log.Err("reinit delete all room user set redis err, %v\n", err)
			}
		}
		err := cacheroom.UpdateRoom(mdr)
		if err != nil {
			log.Err("reinit update room redis err, %v", err)
			continue
		}

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
			log.Err("reinit update room transaction err, %v", err)
			continue
		}
		//读写分离
	}
	return rooms
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
				log.Debug("GiveUpRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n", room.RoomID,
					room.Password, sub.Minutes())
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
				log.Err("room delay room destroy delete room users redis err, %v", err)
				return err
			}
		}
	}
	return nil
}

func DeadRoomDestroy() error {
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
		return nil
	}
	for _, room := range rooms {
		log.Debug("DeadRoomDestroyPolling roomid:%d,pwd:%s\n", room.RoomID, room.Password)
		err := cacheroom.DeleteAllRoomUser(room.Password, "DeadRoomDestroy")
		if err != nil {
			log.Err("delete dead room users redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacheroom.DeleteRoom(room)
		if err != nil {
			log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacheroom.SetRoomDelete(room.GameType, room.RoomID)
		if err != nil {
			log.Err("delete dead set delete room redis err, %d|%v\n", room.RoomID, err)
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
			RoomRefund(room)
		}
		UpdateRoom(room)
		//err = mail.SendSysMail()

	}

	return nil
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

func chekcGameParam(maxNumber int32, maxRound int32, gtype int32, gameParam string) error {
	if len(gameParam) == 0 {
		return errroom.ErrGameParam
	}
	if maxNumber < 2 {
		return errroom.ErrRoomMaxNumber
	}
	if maxRound != 10 && maxRound != 20 && maxRound != 30 {
		return errroom.ErrRoomMaxRound
	}
	//fmt.Printf("ChekcGameParam:%d|%d|%d|%s\n",maxNumber,maxRound,gtype,gameParam)
	switch gtype {
	case enumroom.ThirteenGameType:
		if maxNumber > 4 {
			return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.ThirteenRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			return errroom.ErrGameParam
		}
		//if roomParam.BankerType != 1 && roomParam.BankerType != 2 {
		//	return errors.ErrGameParam
		//}
		if roomParam.BankerAddScore < 0 || roomParam.BankerAddScore > 6 || roomParam.BankerAddScore%2 != 0 {
			return errroom.ErrGameParam
		}
		if roomParam.Joke != 0 && roomParam.Joke != 1 {
			return errroom.ErrGameParam
		}
		if roomParam.Times < 1 || roomParam.Times > 3 {
			return errroom.ErrGameParam
		}
		break
	case enumroom.NiuniuGameType:
		if maxNumber != 6 && maxNumber != 8 && maxNumber != 10 {
			return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.NiuniuRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			return errroom.ErrGameParam
		}
		if roomParam.BankerType < 1 || roomParam.BankerType > 4 {
			return errroom.ErrGameParam
		}
		if roomParam.Times != 3 && roomParam.Times != 5 && roomParam.Times != 10 {
			return errroom.ErrGameParam
		}
		break
	case enumroom.DoudizhuGameType:
		if maxNumber != 4 {
			return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.DoudizhuRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("doudizhu unmarshal room param failed, %v", err)
			return errroom.ErrGameParam
		}
		if roomParam.BaseScore != 0 && roomParam.BaseScore != 5 && roomParam.BaseScore != 10 {
			return errroom.ErrGameParam
		}
		break
	case enumroom.FourCardGameType:
		if maxNumber < 2 && maxNumber > 8 {
			return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.FourCardRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("fourcard unmarshal room param failed, %v", err)
			return errroom.ErrGameParam
		}
		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
			return errroom.ErrGameParam
		}
		if roomParam.BetType < 1 || roomParam.BetType > 2 {
			return errroom.ErrGameParam
		}
		break
	default:
		return errroom.ErrGameParam
	}
	return nil
}

func CheckHasRoom(uid int32) (bool, *mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(uid)
	if hasRoom {
		mdr, err := cacheroom.GetRoomUserID(uid)
		if err != nil {
			return hasRoom, nil, err
		}
		return hasRoom, mdr, err
	}
	return false, nil, nil
}

func GameStart(uid int32) error {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return err
	}
	if uid != mdr.Users[0].UserID {
		return errroom.ErrNotPayer
	}
	if mdr.Status > enumroom.RoomStatusInit {
		return errroom.ErrGameHasBegin
	}
	if len(mdr.Users) < 2 {
		return errroom.ErrPlayerNumberNoEnough
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
		return errroom.ErrRoomNotAllReady
	}
	_, mdPayer := cacheuser.GetUserByID(mdr.PayerID)
	mdr.StartMaxNumber = mdr.MaxNumber
	mdr.MaxNumber = int32(len(mdr.Users))
	RoomRefund(mdr)
	cost := getRoomCost(mdr.GameType, mdr.MaxNumber, mdr.RoundNumber, mdPayer.Channel, mdPayer.Version, mdPayer.MobileOs, mdr.RoomType)
	mdr.Cost = cost
	err = roomStartBalance(mdr, mdPayer)
	if err != nil {
		log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
		return err
	}
	err = cacheroom.UpdateRoom(mdr)
	if err != nil {
		log.Err("room game start failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
		return err
	}
	return nil
}

func RoomCreateBalance(mdr *mdroom.Room, mdu *mduser.User) error {
	//jType := getRoomJournalType(mdr.GameType)
	if mdr.Cost != 0 {
		if mdr.RoomType == enumroom.RoomTypeClub {
			err := club.SetClubBalance(-mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, mdr.GameType*100+1,
				int64(mdr.RoomID), int64(mdu.UserID))
			if err != nil {
				return err
			}
		} else {
			_, err := bill.SetBalanceFreeze(mdu.UserID, int64(mdr.RoomID), &mbill.Balance{Amount: mdr.Cost,
				CoinType: enumcom.Diamond}, mdr.GameType*100+2)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func roomBackUnFreezeAndBalance(mdr *mdroom.Room) error {
	//jType := getRoomJournalType(mdr.GameType)
	if mdr.Cost != 0 {
		if mdr.RoomType != enumroom.RoomTypeClub {
			_, err := bill.SetBalanceFreeze(mdr.PayerID, int64(mdr.RoomID),
				&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond}, mdr.GameType*100+2)
			if err != nil {
				log.Err("back room freeze err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
					mdr.PayerID, -mdr.Cost, mdr.CostType, err)
				return err
			}
			err = bill.GainGameBalance(mdr.PayerID, mdr.RoomID, mdr.GameType*100+1, &mbill.Balance{Amount:
			-mdr.Cost, CoinType: enumcom.Diamond})
			if err != nil {
				log.Err("back room balance err roomid:%d,payerid:%d,cost:%d,constype:%d,err:%v", mdr.RoomID,
					mdr.PayerID, mdr.Cost, mdr.CostType, err)
				return err
			}
		}
	}
	return nil
}

func roomStartBalance(mdr *mdroom.Room, mdu *mduser.User) error {
	//jType := getRoomJournalType(mdr.GameType)
	if mdr.Cost != 0 {
		if mdr.RoomType == enumroom.RoomTypeClub {
			err := club.SetClubBalance(-mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, mdr.GameType*100+1,
				int64(mdr.RoomID), int64(mdu.UserID))
			if err != nil {
				return err
			}
		} else {
			err := bill.GainGameBalance(mdu.UserID, mdr.RoomID, mdr.GameType*100+1,
				&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RoomRefund(mdr *mdroom.Room) error {
	if mdr.Cost != 0 {
		jType := mdr.GameType*100 + 2 //getRoomJournalType(mdr.GameType)
		f := func(tx *gorm.DB) error {
			if mdr.RoomType == enumroom.RoomTypeClub {
				err := club.SetClubBalance(mdr.Cost, enumbill.TypeDiamond, mdr.ClubID, jType, int64(mdr.RoomID),
					int64(mdr.PayerID))
				if err != nil {
					return err
				}
			} else {
				_, err := bill.SetBalanceFreeze(mdr.PayerID, int64(mdr.RoomID),
					&mbill.Balance{Amount: -mdr.Cost, CoinType: enumcom.Diamond}, jType)
				if err != nil {
					log.Err("back room cost err roomid:%d,payerid:%d,cost:%d,constype:%d", mdr.RoomID,
						mdr.PayerID, mdr.Cost, mdr.CostType)
					return err
				}
			}
			return nil
		}
		err := db.Transaction(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func PageSpecialGameList(page *mdpage.PageOption, plgr *mdroom.PlayerSpecialGameRecord) (
	[]*mdroom.PlayerSpecialGameRecord, int64, error) {
	return dbroom.PageSpecialGameList(db.DB(), plgr, page)
}
