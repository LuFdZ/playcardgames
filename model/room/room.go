package room

import (
	"math/rand"
	dbbill "playcards/model/bill/db"
	mdpage "playcards/model/page"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	"playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	mdu "playcards/model/user/mod"
	pbr "playcards/proto/room"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	//"fmt"
)

func GenerateRangeNum(min, max int) string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	return strconv.Itoa(randNum)
}

func RenewalRoom(pwd string, user *mdu.User) ([]int32, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, nil, errors.ErrUserAlreadyInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, errors.ErrRoomNotExisted
	}
	if room.Status != enumr.RoomStatusDelay {
		return nil, nil, errors.ErrRenewalRoon
	}
	ids := room.Ids

	newroom, err := CreateRoom(room.RoomType, room.GameType, room.MaxNumber,
		room.RoundNumber, room.GameParam, user, pwd)
	if err != nil {
		log.Err("room Renewal create new room failed,%v | %v\n", room, err)
		return nil, nil, err
	}
	cacher.DeleteRoom(room.Password)
	err = cacher.SetRoom(newroom)
	if err != nil {
		log.Err("room Renewal set redis failed,%v | %v\n", room, err)
		return nil, nil, err
	}
	return ids, newroom, nil
}

func CreateRoom(rtype int32, gtype int32, maxNum int32, roundNum int32,
	gparam string, user *mdu.User, pwd string) (*mdr.Room,
	error) {

	var err error
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 {
		return nil, errors.ErrUserAlreadyInRoom
	}

	balance, err := dbbill.GetUserBalance(db.DB(), user.UserID)
	if err != nil {
		return nil, err
	}
	if balance.Diamond < int64(maxNum*roundNum*enumr.ThirteenGameCost) {
		return nil, errors.ErrNotEnoughDiamond
	}

	if len(pwd) == 0 {
		pwdNew := GenerateRangeNum(enumr.RoomCodeMin, enumr.RoomCodeMax)
		exist, err := cacher.CheckRoomExist(pwdNew)
		for i := 0; exist && i < 3; i++ {
			pwdNew = GenerateRangeNum(enumr.RoomCodeMin, enumr.RoomCodeMax)
			exist, err = cacher.CheckRoomExist(pwdNew)
			if err != nil {
				return nil, err
			}
		}

		if exist {
			return nil, errors.ErrRoomPwdExisted
		}
		pwd = pwdNew
	}
	users := []*mdr.RoomUser{}
	if rtype == 0 {
		rtype = 1
	}
	if rtype == enumr.RoomTypeNom {

		roomUser := GetRoomUser(user, enumr.UserUnready, 1,
			enumr.UserRoleMaster)
		users = append(users, roomUser)
	}

	now := gorm.NowFunc()
	room := &mdr.Room{
		Password:    pwd,
		GameType:    gtype,
		MaxNumber:   maxNum,
		RoundNumber: roundNum,
		RoundNow:    1,
		GameParam:   gparam,
		Status:      enumr.RoomStatusInit,
		Giveup:      enumr.NoGiveUp,
		Users:       users,
		RoomType:    rtype,
		PayerID:     user.UserID,
		GiveupAt:    &now,
		CreatedAt:	 &now,
		UpdatedAt:	 &now,
		Ids:         []int32{user.UserID},
	}

	f := func(tx *gorm.DB) error {
		err := dbr.CreateRoom(tx, room)
		if err != nil {
			log.Err("room create failed,%v | %v", room, err)
			return err
		}
		err = cacher.SetRoom(room)
		if err != nil {
			log.Err("room create set redis failed,%v | %v\n", room, err)
			return err
		}
		err =cacher.SetRoomUser(room.RoomID, room.Password, user.UserID)
		if err != nil {
			log.Err("room create set room user set redis failed,%v | %v\n", room, err)
			return err
		}
		return nil
	}
	go db.Transaction(f)

	//读写分离
	//f := func(tx *gorm.DB) error {
	//	err := dbr.CreateRoom(tx, room)
	//	if err != nil {
	//		log.Err("room create failed,%v | %v", room, err)
	//		return err
	//	}
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	cacher.DeleteRoom(room.Password)
	//	return nil, err
	//}
	//读写分离

	return room, nil

}

func JoinRoom(pwd string, user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, nil, errors.ErrUserAlreadyInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, errors.ErrRoomNotExisted
	}
	if room.Status != enumr.RoomStatusInit {
		return nil, nil, errors.ErrNotReadyStatus
	}
	if room.Giveup == enumr.WaitGiveUp {
		return nil, nil, errors.ErrInGiveUp
	}
	num := len(room.Users)
	if num >= (int)(room.MaxNumber) {
		return nil, nil, errors.ErrRoomFull
	}

	for i, roomuser := range room.Users {
		if roomuser.UserID == user.UserID {
			return nil, nil, errors.ErrUserAlreadyInRoom
		}
		if roomuser.Position != int32(i+1) {
			num = i
		}
	}
	roomUser := GetRoomUser(user, enumr.UserUnready, int32(num+1),
		enumr.UserRoleSlave)
	newUsers := append(room.Users, roomUser)
	room.Users = newUsers
	room.Ids = append(room.Ids,user.UserID)
	err = cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v|%v\n", err, room)
		return nil, nil, err
	}
	err = cacher.SetRoomUser(room.RoomID, room.Password, user.UserID)
	if err != nil {
		log.Err("room user jooin set session failed, %v|%v\n", err, room)
		return nil, nil, err
	}
	//读写分离
	//f := func(tx *gorm.DB) error {
	//	_, err := dbr.UpdateRoom(tx, room)
	//	if err != nil {
	//		log.Err("room jooin db failed, %v", err)
	//		return err
	//	}
	//	//room = r
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return nil, nil, err
	//}
	//读写分离

	return roomUser, room, nil
}

func LeaveRoom(user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	pwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(pwd) == 0 {
		return nil, nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}

	if room.RoundNumber > 1 && room.Status > enumr.RoomStatusAllReady {
		return nil, nil, errors.ErrGameHasBegin
	}

	if room.Giveup == enumr.WaitGiveUp {
		return nil, nil, errors.ErrInGiveUp
	}

	newUsers := []*mdr.RoomUser{}
	temp := []*mdr.RoomUser{}
	roomUser := &mdr.RoomUser{}
	handle := 0
	isDestroy := 0
	for i := range room.Users {
		if room.Users[i].UserID == user.UserID &&
			room.Status == enumr.RoomStatusInit {
			handle = 1
			roomUser = room.Users[i]
			//若离开的是房主，解散房间
			if room.Users[i].Role == enumr.UserRoleMaster {
				log.Info("delete room cause master leave.user:%d,room:%d\n",
					user.UserID, room.RoomID)
				isDestroy = 1
				//room.Users = nil
				room.Status = enumr.RoomStatusDestroy
				break
			}
		} else {
			temp = newUsers
			newUsers = append(temp, room.Users[i])
		}
	}
	var ids []int32
	for _,user :=range room.Users{
		ids = append(ids,user.UserID)
	}
	room.Ids = ids
	if handle == 0 {
		return nil, nil, errors.ErrUserNotInRoom
	}
	room.Users = newUsers

	//读写分离
	//f := func(tx *gorm.DB) error {
	//	_, err := dbr.UpdateRoom(tx, room)
	//	if err != nil {
	//		return err
	//	}
	//	room = r
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return nil, nil, err
	//}
	//读写分离

	if isDestroy == 1 || len(newUsers) == 0 {
		err = cacher.DeleteAllRoomUser(room.Password,"LeaveRoom")
		if err != nil {
			log.Err("leave room delete all users failed, %d|%v\n",
				room.RoomID, err)
		}
		err = cacher.DeleteRoom(room.Password)
		if err != nil {
			log.Err("leave room delete room failed, %d|%v\n",
				room.RoomID, err)
		}
	} else {
		err :=cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room leave room delete failed, %d|%v\n",
				room.RoomID, err)
		}
		err = cacher.DeleteRoomUser(user.UserID)
		if err != nil {
			log.Err("room leave room delete user failed, %d|%v\n",
				room.RoomID, err)
		}
	}

	return roomUser, room, nil
}

func GetReady(pwd string, uid int32) (*mdr.RoomUser,[]int32, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil,nil,errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil,nil,err
	}
	if room.Status > enumr.RoomStatusInit {
		return nil,nil,errors.ErrNotReadyStatus
	}
	if room.Giveup == enumr.WaitGiveUp {
		return nil,nil,errors.ErrInGiveUp
	}
	allReady := true
	t := time.Now()
	var num int32
	for _, user := range room.Users {
		num++
		if user.UserID == uid {
			if user.Ready == enumr.UserReady {
				return nil,nil,errors.ErrAlreadyReady
			}
			user.Ready = enumr.UserReady
			user.UpdatedAt = &t
			//out = user

		}
		if allReady && user.Ready != enumr.UserReady {
			allReady = false
		}
	}
	if allReady && num == room.MaxNumber {
		room.Status = enumr.RoomStatusAllReady
	}

	err =cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room ready failed, %d|%d|%v\n",room.RoomID,uid, err)
	}
	//读写分离
	//f := func(tx *gorm.DB) error {
	//	r, err := dbr.UpdateRoom(tx, room)
	//	if err != nil {
	//		return err
	//	}
	//	room = r
	//	return nil
	//}
	//err = db.Transaction(f)
	//if err != nil {
	//	return nil, 0, err
	//}
	//读写分离

	readyUser := &mdr.RoomUser{
		UserID: uid,
	}

	return readyUser,room.Ids, nil
}

//func GetRoomsByStatus(status int32) ([]*mdr.Room, error) {
//	var (
//		rooms []*mdr.Room
//	)
//	f := func(tx *gorm.DB) error {
//		list, err := dbr.GetRoomsByStatus(db.DB(), status)
//		if err != nil {
//			return err
//		}
//		rooms = list
//		return nil
//	}
//	err := db.Transaction(f)
//	if err != nil {
//		return nil, err
//	}
//	return rooms, nil
//}

func GiveUpGame(pwd string, uid int32) ([]int32,*mdr.GiveUpGameResult,
	error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil,nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil,nil, err
	}

	if room.RoundNumber == 1 && room.Status < enumr.RoomStatusStarted {
		return nil,nil, errors.ErrNotReadyStatus
	}

	if room.Giveup == enumr.WaitGiveUp {
		return GiveUpVote(pwd, 1, uid)
	}

	giveUpResult := room.GiveupGame
	var list []*mdr.UserState
	agreeGiveUp := 0
	for _, user := range room.Users {
		var state int32
		scoketstatus := cacher.GetUserStatus(user.UserID)
		if user.UserID == uid {
			state = enumr.UserStateLeader
			agreeGiveUp++
		} else if scoketstatus == enumr.SocketClose {
			state = enumr.UserStateOffline
			agreeGiveUp++
		} else {
			state = enumr.UserStateWaiting
		}
		us := &mdr.UserState{
			UserID: user.UserID,
			State:  state,
		}
		list = append(list, us)
	}
	if agreeGiveUp == len(room.Users) {
		room.Status = enumr.RoomStatusGiveUp
		giveUpResult.Status = enumr.GiveupStatusAgree
	} else {
		room.Giveup = enumr.WaitGiveUp
		now := gorm.NowFunc()
		room.GiveupAt = &now
		giveUpResult.Status = enumr.GiveupStatusWairting
	}
	//giveUpResult.RoomID = room.RoomID

	giveUpResult.UserStateList = list
	room.GiveupGame = giveUpResult
	if room.Status == enumr.RoomStatusGiveUp {
		err = cacher.DeleteAllRoomUser(room.Password,"GiveUpGame")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil,nil, err
		}
		err = cacher.DeleteRoom(room.Password)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil,nil, err
		}
	} else {
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil,nil, err
		}
	}
	//读写分离
	//_, err = UpdateRoom(room)
	//if err != nil {
	//	log.Err("room give up update failed, %v", err)
	//	return nil, err
	//}
	//读写分离

	log.Info("give up game:%d|%v", uid, giveUpResult)
	return room.Ids,&giveUpResult, nil
}

func GiveUpVote(pwd string, status int32, uid int32) ([]int32,*mdr.GiveUpGameResult,
	error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil,nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil,nil, err
	}

	if room.Giveup != enumr.WaitGiveUp {
		return nil,nil, errors.ErrNotInGiveUp
	}

	giveUpResult := room.GiveupGame
	if status == 1 {
		status = enumr.UserStateAgree
	} else {
		status = enumr.UserStateDisagree
	}
	giveup := enumr.GiveupStatusWairting
	agreeGiveUp := 0
	for _, userstate := range room.GiveupGame.UserStateList {
		if userstate.UserID == uid {
			if userstate.State != enumr.UserStateWaiting {
				return nil,nil, errors.ErrAlreadyVoted
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
	if agreeGiveUp == len(room.GiveupGame.UserStateList) {
		room.Status = enumr.RoomStatusGiveUp
		giveUpResult.Status = enumr.GiveupStatusAgree
		err = cacher.DeleteAllRoomUser(room.Password,"GiveUpVote")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return nil,nil, err
		}
		err = cacher.DeleteRoom(room.Password)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil,nil, err
		}
	} else if giveup == enumr.GiveupStatusDisAgree {
		room.Giveup = enumr.NoGiveUp
		giveUpResult.Status = enumr.GiveupStatusDisAgree
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil,nil, err
		}
	} else {
		room.Giveup = enumr.WaitGiveUp
		giveUpResult.Status = enumr.GiveupStatusWairting
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil,nil, err
		}
	}

	//读写分离
	//_, err = UpdateRoom(room)
	//if err != nil {
	//	log.Err("room give up update failed, %v", err)
	//	return nil, err
	//}
	//fmt.Printf("give up game:%v", giveUpResult)
	//读写分离

	log.Info("give up game vote:%d|%v", uid, giveUpResult)
	return room.Ids,&giveUpResult, nil
}

func GetRoomUser(u *mdu.User, ready int32, position int32,
	role int32) *mdr.RoomUser {
	now := gorm.NowFunc()
	return &mdr.RoomUser{
		UserID:    u.UserID,
		Ready:     ready,
		Nickname:  u.Nickname,
		Position:  position,
		Icon:      u.Icon,
		Sex:       u.Sex,
		Role:      role,
		UpdatedAt: &now,
	}
}

func Shock(uid int32, sendid int32) ( error) {
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

	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return err
	}
	if room.Status != enumr.RoomStatusStarted {
		return errors.ErrNotReadyStatus
	}

	return nil
}

func VoiceChat(uid int32) (*mdr.Room, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if room.Status > enumr.RoomStatusDone {
		return nil, errors.ErrGameIsDone
	}
	return room, nil
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

func CreateFeedback(fb *mdr.Feedback) (*mdr.Feedback, error) {
	return dbr.CreateFeedback(db.DB(), fb)
}

func RoomResultList(uid int32, gtype int32) (*pbr.RoomResultListReply, error) {
	var list []*pbr.RoomResults
	rooms, err := dbr.GetRoomResultByUserIdAndGameType(db.DB(), uid, gtype)
	if err != nil {
		return nil, err
	}
	for _, room := range rooms {
		result := &mdr.RoomResults{
			Status:    room.Status,
			Password:  room.Password,
			GameType:  room.GameType,
			CreatedAt: room.CreatedAt,
			List:      room.UserResults,
		}
		list = append(list, result.ToProto())
	}
	out := &pbr.RoomResultListReply{
		List: list,
	}
	return out, nil
}

func CheckRoomExist(uid int32) (int32, *mdr.CheckRoomExist, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	//fmt.Printf("AAAAA CRE:%s", pwd)
	if len(pwd) == 0 {
		log.Err("CheckRoomExistPWDNULL:%s|%d",pwd,uid)
		return 2, nil, nil
	}

	room, err := cacher.GetRoom(pwd)
	//fmt.Printf("BBBB CRE:%s|%v", pwd, room)
	if err != nil {
		return 2, nil, err
	}

	if room == nil {
		//log.Err("check pwd exist but room null, %s|%d", pwd, uid)
		err = cacher.DeleteAllRoomUser(pwd,"CheckRoomExistRoomNull")
		if err != nil {
			log.Err("room give up delete room users redis err, %v", err)
			return 2, nil, err
		}
		log.Err("CheckRoomExistROOMNULL:%s|%d",pwd,uid)
		return 2, nil, nil
	} else if room.Status > enumr.RoomStatusDelay {
		err = cacher.DeleteAllRoomUser(room.Password,"CheckRoomExistRoomDelay")
		if err != nil {
			log.Err("check room delete room users redis err, %v", err)
			return 2, nil, err
		}
		err = cacher.DeleteRoom(pwd)
		if err != nil {
			log.Err("check room delete redis err, %s|%v", pwd, err)
			return 2, nil, err
		}
	}

	var roomStatus int32
	if room.Status == enumr.RoomStatusInit {
		for _, roomuser := range room.Users {
			if roomuser.UserID == uid {
				if roomuser.Ready == enumr.UserUnready {
					if room.RoundNow == 1 {
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
		//return 2, nil, nil
	} else if room.Status == enumr.RoomStatusReInit {
		roomStatus = enumr.RecoveryInitNoReady
	} else {
		roomStatus = enumr.RecoveryGameStart
	}
	Results := mdr.RoomResults{
		RoundNumber: room.RoundNumber,
		RoundNow:    room.RoundNow,
		Status:      room.Status,
		Password:    room.Password,
		List:        room.UserResults,
	}
	roomResults := &mdr.CheckRoomExist{
		Result:       1,
		Room:         *room,
		Status:       roomStatus,
		GiveupResult: room.GiveupGame,
		GameResult:   Results,
	}

	return 1, roomResults, nil
}

func RoomUserStatusCheck() []*pbr.UserConnection {
	var ucs []*pbr.UserConnection
	//rooms, err := GetRoomsByStatus(enumr.RoomStatusStarted)
	//if err != nil {
	//	log.Err("room user status check err:%v", err)
	//	return nil
	//}
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
	for _, room := range rooms {
		//Room.UserOffAline
		for _, user := range room.Users {
			status := cacher.GetUserStatus(user.UserID)
			notice := cacher.GetUserSocketNotice(user.UserID)

			//若房间不在游戏开始状态 或者连接状态未初始化 或者已广播过 不做处理
			if (room.Status != enumr.RoomStatusInit &&
				room.Status != enumr.RoomStatusStarted &&
				room.Status != enumr.RoomStatusAllReady &&
				room.Status != enumr.RoomStatusReInit) ||
				status == 0 || notice == 1 {
				continue
			}

			uc := &pbr.UserConnection{
				Ids: room.Ids,
				UserID: user.UserID,
				Status: status,
			}

			if status == enumr.SocketClose {
				cacher.UpdateRoomUserSocektStatus(user.UserID, enumr.SocketClose, 1)
			} else if status == enumr.SocketAline {
				cacher.UpdateRoomUserSocektStatus(user.UserID, enumr.SocketAline, 1)
			}
			ucs = append(ucs, uc)
		}
	}
	return ucs
}

func ReInit() []*mdr.Room {
	//rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusReInit)
	//if err != nil {
	//	log.Err("reinit get rooms by status err, %v\n", err)
	//	return nil
	//}

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
			err := cacher.DeleteAllRoomUser(room.Password,"ReInitRoomDelay")
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
		go db.Transaction(f)
		//读写分离
		//err = db.Transaction(f)
		//if err != nil {
		//	log.Err("reinit update room transaction err, %v", err)
		//	continue
		//}
		//读写分离
	}
	return rooms
}

func GiveUpRoomDestroy() []*mdr.Room {
	//rooms, err := dbr.GetRoomsGiveup(db.DB())
	//if err != nil {
	//	log.Err("giveup room destroy get rooms by status err, %v", err)
	//	return nil
	//}

	//tx.Where(" giveup = ? and status < ?", enumr.WaitGiveUp,
	//	enumr.RoomStatusDone)

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
				err = cacher.DeleteAllRoomUser(room.Password,"GiveUpRoomDestroy")
				if err != nil {
					log.Err("room give up destroy delete room users redis err, %v", err)
					continue
				}
				err = cacher.DeleteRoom(room.Password)
				if err != nil {
					log.Err("room give up destroy delete room redis err, %v", err)
					continue
				}
				err = cacher.SetRoomDelete(room.GameType,room.RoomID)
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
				log.Debug("GiveUpRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n",room.RoomID,room.Password,sub.Minutes())
				go db.Transaction(f)
				//读写分离
				//err = db.Transaction(f)
				//if err != nil {
				//	log.Err("room give up destroy delete room users redis err, %v", err)
				//	continue
				//}
				//读写分离

				giveRooms = append(giveRooms, checkRoom)
			}
		}
	}
	return giveRooms
}

func DelayRoomDestroy() error {
	//rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusDelay)
	//if err != nil {
	//	log.Err("reinit get rooms by status err, %v", err)
	//	return nil
	//}
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
				err = cacher.DeleteAllRoomUser(room.Password,"DelayRoomDestroy")
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
			err = cacher.SetRoomDelete(room.GameType,room.RoomID)
			if err != nil {
				log.Err("delay set delete room redis err, %d|%v\n", room.RoomID, err)
			}
			room.Status = enumr.RoomStatusDone
			log.Debug("DelayRoomDestroyPolling roomid:%d,pwd:%s,subdate:%f m\n",room.RoomID,room.Password,sub.Minutes())
			f := func(tx *gorm.DB) error {
				_, err := dbr.UpdateRoom(tx, room)
				if err != nil {
					log.Err("room delay destroy room db err, %v\n", err)
					return err
				}
				//room = r
				return nil
			}
			go db.Transaction(f)
			//err = db.Transaction(f)
			//if err != nil {
			//	return err
			//}
		}
	}
	return nil
}

func DeadRoomDestroy() error {
	//定时清除不活动的房间
	f := func(r *mdr.Room) bool {
		sub := time.Now().Sub(*r.UpdatedAt)
		//fmt.Printf("DeadRoomDestroy:%d\n",sub.Hours())
		if sub.Hours()>24 {
			log.Debug("DeadRoomDestroyDate roomid:%d,pwd:%s,subdate:%f h\n",r.RoomID,r.Password,sub.Hours())
			return true
		}
		return false
	}
	rooms := cacher.GetAllRoom(f)
	if len(rooms) == 0 {
		return nil
	}
	for _, room := range rooms {
		log.Debug("DeadRoomDestroyPolling roomid:%d,pwd:%s\n",room.RoomID,room.Password)
		err := cacher.DeleteAllRoomUser(room.Password,"DeadRoomDestroy")
		if err != nil {
			log.Err("delete dead room users redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacher.DeleteRoom(room.Password)
		if err != nil {
			log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
		}
		err = cacher.SetRoomDelete(room.GameType,room.RoomID)
		if err != nil {
			log.Err("delete dead set delete room redis err, %d|%v\n", room.RoomID, err)
		}
		room.Status = room.Status*10+enumr.RoomStatusOverTimeClean
		f := func(tx *gorm.DB) error {
			_,err = dbr.UpdateRoom(tx,room)
			if err != nil {
				log.Err("delete dead room redis err, %d|%v\n", room.RoomID, err)
			}
			return nil
		}
		go db.Transaction(f)
	}

	return nil
}
