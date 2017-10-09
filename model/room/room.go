package room

import (
	"fmt"
	"math/rand"
	dbbill "playcards/model/bill/db"
	mdpage "playcards/model/page"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	mdt "playcards/model/thirteen/mod"
	mdu "playcards/model/user/mod"
	pbr "playcards/proto/room"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	lua "github.com/yuin/gopher-lua"
)

func GenerateRangeNum(min, max int) string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	return strconv.Itoa(randNum)
}

func RenewalRoom(pwd string, user *mdu.User) (int32, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return 0, nil, errors.ErrUserAlreadyInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, nil, err
	}
	if room == nil {
		return 0, nil, errors.ErrRoomNotExisted
	}
	if room.Status != enumr.RoomStatusDelay {
		return 0, nil, errors.ErrRenewalRoon
	}
	oldID := room.RoomID
	cacher.DeleteRoom(room.Password)
	room, err = CreateRoom(room.RoomType, room.GameType, room.MaxNumber,
		room.RoundNumber, room.GameParam, user, pwd)

	err = cacher.SetRoom(room)
	if err != nil {
		log.Err("room create set redis failed,%v | %v", room, err)
		return 0, nil, err
	}
	return oldID, room, nil
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

	if balance.Diamond < int64(roundNum*enumr.ThirteenGameCost/10) {
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

	//users := []*mdr.RoomUser{roomUser}

	room := &mdr.Room{
		Password:    pwd,
		GameType:    gtype,
		MaxNumber:   maxNum,
		RoundNumber: roundNum,
		RoundNow:    1,
		GameParam:   gparam,
		//UserList:  fmt.Sprint(userID),
		Status:   enumr.RoomStatusInit,
		Users:    users,
		RoomType: rtype,
		PayerID:  user.UserID,
	}

	f := func(tx *gorm.DB) error {
		err := dbr.CreateRoom(tx, room)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		cacher.DeleteRoom(room.Password)
		return nil, err
	}

	err = cacher.SetRoom(room)
	if err != nil {
		log.Err("room create set redis failed,%v | %v", room, err)
		return nil, err
	}

	cacher.SetRoomUser(room.RoomID, room.Password, user.UserID)
	//webroom.AutoSubscribe(u.UserID)
	return room, nil

}

func JoinRoom(pwd string, user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, nil, errors.ErrUserAlreadyInRoom
	}
	fmt.Printf("AAA Join Room:%s", checkpwd)
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, errors.ErrRoomNotExisted
	}
	num := len(room.Users)
	if num >= (int)(room.MaxNumber) {
		return nil, nil, errors.ErrRoomFull
	}
	//var position int32
	// isOrder := true
	// for index, roomUser := range room.Users {
	// 	if int32(index+1) != roomUser.Position {
	// 		isOrder = false
	// 		position = index + 1
	// 		break
	// 	}
	// }
	// if isOrder {
	// 	position = num + 1
	// }

	for _, roomuser := range room.Users {
		if roomuser.UserID == user.UserID {
			return nil, nil, errors.ErrUserAlreadyInRoom
		}
	}

	roomUser := GetRoomUser(user, enumr.UserUnready, int32(num+1),
		enumr.UserRoleSlave)
	newUsers := append(room.Users, roomUser)
	room.Users = newUsers
	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		room = r
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, nil, err
	}

	err = cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, nil, err
	}
	cacher.SetRoomUser(room.RoomID, room.Password, user.UserID)
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

	if room.Status == enumr.RoomStatusAllReady {
		return nil, nil, errors.ErrGameHasBegin
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
				log.Info("delete room cause master leave.user:%d,room:%d",
					user.UserID, room.RoomID)
				isDestroy = 1
				//room.Users = nil
				break
			}
		} else {
			temp = newUsers
			newUsers = append(temp, room.Users[i])
		}
	}
	if handle == 0 {
		return nil, nil, errors.ErrUserNotInRoom
	}
	cacher.DeleteRoomUser(user.UserID)
	if isDestroy == 1 || len(newUsers) == 0 {
		room.Status = enumr.RoomStatusDestroy
		cacher.DeleteRoom(room.Password)
	} else {
		cacher.UpdateRoom(room)
	}

	room.Users = newUsers
	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		room = r
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, nil, err
	}

	//cacheroom.DeleteRoomUser(room.RoomID, user.UserID)
	// if err != nil {
	// 	log.Err("room leave set session failed, %v", err)
	// 	return nil, nil, err
	// }
	err = cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room leave room set session failed, %v", err)
		return nil, nil, err
	}
	return roomUser, room, nil
}

//GetReadyOrUnReady, readyStatus int32
func GetReady(pwd string, uid int32) (*mdr.RoomUser, int32, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	//fmt.Printf("AAAAAAA: %d", checkpwd)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, 0, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, 0, err
	}
	if room.Status > enumr.RoomStatusInit {
		return nil, 0, errors.ErrNotReadyStatus
	}
	allReady := true
	t := time.Now()
	//out := &mdr.RoomUser{}
	var num int32
	for _, user := range room.Users {
		num++
		if user.UserID == uid {
			if user.Ready == enumr.UserReady {
				return nil, 0, errors.ErrAlreadyReady
			}
			user.Ready = enumr.UserReady
			user.UpdatedAt = &t
			//out = user

		}
		if allReady && user.Ready != enumr.UserReady {
			allReady = false
		}
	}
	//fmt.Printf("CCCCCCC: %t|%d", allReady, num)
	if allReady && num == room.MaxNumber {
		room.Status = enumr.RoomStatusAllReady
	}

	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		room = r
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, 0, err
	}

	err = cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, 0, err
	}
	readyUser := &mdr.RoomUser{
		UserID: uid,
	}
	return readyUser, room.RoomID, nil
}

func GetRoomByUserID(uid int32, pwd string, InOrNotIn bool) (*mdr.Room, error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if InOrNotIn {
		if len(checkpwd) == 0 {
			return nil, errors.ErrUserNotInRoom
		}

	} else {
		if len(pwd) != 0 {
			return nil, errors.ErrUserAlreadyInRoom
		}
	}

	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func GetRoomsByStatus(status int32) ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	f := func(tx *gorm.DB) error {
		list, err := dbr.GetRoomsByStatus(db.DB(), status)
		if err != nil {
			return err
		}
		rooms = list
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func RoomAllReady() error {

	rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusAllReady)
	if err != nil {
		return err
	}
	if len(rooms) == 0 {
		return nil
	}
	var ids []int32
	for _, room := range rooms {
		ids = append(ids, room.RoomID)
		room.Status = enumr.RoomStatusStarted
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room id:%d update status err:%v",
				room.RoomID, err)
			continue
		}
	}

	f := func(tx *gorm.DB) error {

		err = dbr.BatchUpdate(tx,
			enumr.RoomStatusStarted, ids)

		if err != nil {
			log.Err("room update set session failed, %v", err)
			return err
		}

		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return err
	}
	return nil
}

func ReInit() []*mdr.Room {
	rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusReInit)
	if err != nil {
		log.Err("reinit get rooms by status err, %v", err)
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	//	var doneRooms []*mdr.Room
	for _, room := range rooms {
		//房间数局
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态
		//roundNow := room.RoundNow + 1
		fmt.Printf("rooms round now:%d|%d", room.RoundNow, room.RoundNumber)
		if room.RoundNow == room.RoundNumber {
			room.Status = enumr.RoomStatusDelay
			fmt.Printf("rooms status now:%d", room.Status)
		} else {
			room.Status = enumr.RoomStatusInit
			room.RoundNow++
			for _, user := range room.Users {
				user.Ready = enumr.UserUnready
			}
		}

		f := func(tx *gorm.DB) error {
			//更新玩家游戏局数
			dbr.UpdateRoomPlayTimes(tx, room.RoomID, room.GameType)
			r, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				return err
			}
			room = r
			//fmt.Printf("rooms status now:%d", room.Status)
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			log.Err("reinit update room transaction err, %v", err)
			continue
		}

		//fmt.Printf("游戏正常结算后 清除玩家缓存A %d \n", room.Status)
		if room.Status == enumr.RoomStatusDelay {
			//err = cacheroom.DeleteRoom(room.Password)
			//游戏正常结算后 先清除玩家缓存 保留房间缓存做续费重开
			err = cacher.DeleteAllRoomUser(room.Password)
		}
		err = cacher.SetRoom(room)
		// err = cacher.DeleteAllRoomUser(room.Password)
		// err = cacher.SetRoom(room)
		if err != nil {
			log.Err("reinit update room redis err, %v", err)
			continue
		}
	}

	return rooms
}

func GiveUpGame(pwd string, status int32, uid int32) (*mdr.GiveUpGameResult,
	error) {
	checkpwd := cacher.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}

	if room.Status != enumr.RoomStatusStarted &&
		room.Status != enumr.RoomStatusReInit &&
		room.Status != enumr.RoomStatusWaitGiveUp {
		return nil, errors.ErrNotReadyStatus
	}
	//fmt.Printf("GiveUpGame:%d|%d", room.Status, enumr.RoomStatusStarted)

	if status == 1 {
		status = enumr.AgreeGiveUpRoom
	} else {
		status = enumr.DisAgreeGiveUpRoom
	}
	giveUpResult := &room.GiveupGame
	//fmt.Printf("give up game AAAAA:%d", room.Status)
	if room.Status != enumr.RoomStatusWaitGiveUp {
		//fmt.Printf("give up game BBBBB:%d", uid)
		var list []*mdr.UserState
		var giveUpResult mdr.GiveUpGameResult
		for _, user := range room.Users {
			var state int32
			scoketstatus := cacher.GetUserStatus(user.UserID)
			if user.UserID == uid {
				state = enumr.UserStateLeader
			} else if scoketstatus == enumr.SocketClose {
				state = enumr.UserStateOffline
			} else {
				state = enumr.UserStateWaiting
			}
			us := &mdr.UserState{
				UserID: user.UserID,
				State:  state,
			}
			list = append(list, us)
		}

		room.Status = enumr.RoomStatusWaitGiveUp
		giveUpResult.RoomID = room.RoomID
		giveUpResult.Status = room.Status
		giveUpResult.UserStateList = list
		room.GiveupGame = giveUpResult
		//fmt.Printf("give up game BBBBB:%v", giveUpResult)
		err = cacher.UpdateRoom(room)
		if err != nil {
			log.Err("room give up set session failed, %v", err)
			return nil, err
		}

		_, err = UpdateRoom(room)
		if err != nil {
			log.Err("room give up update failed, %v", err)
			return nil, err
		}

	} else {
		giveup := enumr.AgreeGiveUpRoom
		for _, userstate := range room.GiveupGame.UserStateList {
			if userstate.UserID == uid {
				if userstate.State != enumr.UserStateWaiting {
					return nil, errors.ErrAlreadyVoted
				}
				userstate.State = status
				//fmt.Printf("give up game CCCCC:%d", userstate.State)
				if userstate.State == enumr.DisAgreeGiveUpRoom {
					giveup = enumr.DisAgreeGiveUpRoom
					break
				}
			}
			if userstate.State == enumr.UserStateWaiting {
				giveup = enumr.DisAgreeGiveUpRoom
				break
			}
		}
		if giveup == enumr.AgreeGiveUpRoom {
			room.Status = enumr.RoomStatusGiveUp
			giveUpResult.Status = room.Status
			cacher.DeleteRoom(room.Password)
			if err != nil {
				log.Err("room give up set session failed, %v", err)
				return nil, err
			}
			_, err = UpdateRoom(room)
			if err != nil {
				log.Err("room give up update failed, %v", err)
				return nil, err
			}
		}
	}
	fmt.Printf("give up game:%v", giveUpResult)
	return giveUpResult, nil
}

func RoomDestroy() error {

	rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusDelay)
	if err != nil {
		log.Err("reinit get rooms by status err, %v", err)
		return nil
	}
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
				err = cacher.DeleteRoom(room.Password)
				if err != nil {
					log.Err("room destroy delete room redis err, %v", err)
					continue
				}
			}
			room.Status = enumr.RoomStatusDone
			f := func(tx *gorm.DB) error {
				r, err := dbr.UpdateRoom(tx, room)
				if err != nil {
					return err
				}
				room = r
				return nil
			}
			err = db.Transaction(f)
			if err != nil {
				return err
			}
		}

	}

	//定时清除不活动的房间
	f := func(tx *gorm.DB) error {
		pwdList, err := dbr.GetDeadRoomPassword(tx)
		if err != nil {
			return err
		}

		for _, pwd := range pwdList {
			err = cacher.DeleteRoom(pwd)
			if err != nil {
				log.Err("reinit update room redis err, %v", err)
				continue
			}
		}

		err = dbr.CleanDeadRoomByUpdateAt(tx)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		log.Err("reinit update room transaction err, %v", err)
	}
	return nil
}

// func RoomFinish(pwd string, status int32) error {
// 	room, err := cacheroom.GetRoom(pwd)
// 	if err != nil {
// 		return err
// 	}
// 	room.Status = status
// 	f := func(tx *gorm.DB) error {
// 		_, err := dbr.UpdateRoom(tx, room)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// 	err = db.Transaction(f)
// 	if err != nil {
// 		return err
// 	}

// 	err = cacheroom.DeleteRoom(pwd)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func Heartbeat(uid int32) error {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return err
	}
	for _, item := range room.Users {
		if item.UserID == uid {
			t := time.Now()
			item.UpdatedAt = &t
		}
	}

	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		room = r
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return err
	}

	err = cacher.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return err
	}
	return nil
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

func Shock(uid int32, sendid int32) (int32, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	pwdCheck := cacher.GetRoomPasswordByUserID(sendid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}

	if pwd != pwdCheck {
		return 0, errors.ErrNotInSameRoon
	}

	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, err
	}
	if room.Status != enumr.RoomStatusStarted {
		return 0, errors.ErrNotReadyStatus
	}

	return room.RoomID, nil
}

func LuaTest() error {

	l := lua.NewState()
	defer l.Close()
	if err := l.DoFile("lua/Logic.lua"); err != nil {
		fmt.Printf("LuaTest:%v", err)
		return err
	}

	// if err := l.DoString("return Logic:new()"); err != nil {
	// 	fmt.Printf("AAAAA:%v", err)
	// }
	// //logic := l.Get(1)
	// l.Pop(1)

	// if err := l.DoString("return Logic:GetCards()"); err != nil {
	// 	log.Err("thirteen logic do string %v", err)
	// }
	// //getCards := l.Get(1)
	// l.Pop(1)

	// var cardList []string
	// if cardsMap, ok := getCards.(*lua.LTable); ok {
	// 	cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
	// 		if cards, ok := value.(*lua.LTable); ok {
	// 			var cardType string
	// 			var cardValue string
	// 			cards.ForEach(func(k lua.LValue, v lua.LValue) {
	// 				//value, _ := strconv.ParseInt(v.String(), 10, 32)
	// 				if k.String() == "_type" {
	// 					cardType = v.String()
	// 				} else {
	// 					cardValue = v.String()
	// 				}
	// 			})
	// 			fmt.Printf("AAAAAACardMapsValue : %s|%s\n", cardType, cardValue)
	// 			// card := mdt.Card{
	// 			// 	Type:  int32(cardType),
	// 			// 	Value: int32(cardValue),
	// 			// }

	// 			cardList = append(cardList, cardType+"_"+cardValue)
	// 		} else {
	// 			log.Err("thirteen cardsMap value err %v", value)
	// 		}
	// 	})
	// } else {
	// 	log.Err("thirteen cardsMap err %v", cardsMap)
	// }
	//testGetCards
	test := "{\"a\" : \"1\"}"
	if err := l.DoString(fmt.Sprintf("return Logic:GetResult('%s')", test)); err != nil {
		fmt.Printf("DoString GetResult:%v", err)
	}

	logic := l.Get(1)
	// l.Pop(1)

	var gameresult mdt.GameResultList
	var tResults []*mdt.ThirteenResult
	if test, ok := logic.(*lua.LTable); ok {
		test.ForEach(func(key lua.LValue, value lua.LValue) {
			if key.String() == "RoomID" {
				rid, _ := strconv.ParseInt(value.String(), 10, 32)
				gameresult.RoomID = int32(rid)
			} else if key.String() == "ResultList" {
				if results, ok := value.(*lua.LTable); ok {
					results.ForEach(func(k lua.LValue, v lua.LValue) {
						var tResult *mdt.ThirteenResult
						if k.String() == "UserID" {
							uid, _ := strconv.ParseInt(v.String(), 10, 32)
							tResult.UserID = int32(uid)
						} else if k.String() == "Settle" {
							if settles, ok := value.(*lua.LTable); ok {
								settles.ForEach(func(h lua.LValue, s lua.LValue) {
									if h.String() == "Head" {

									}
								})
							}
						} else if k.String() == "Result" {

						}

						tResults = append(tResults, tResult)
					})
				} else {
					fmt.Printf("thirteen settle value err %s|%v", key, value)
					log.Err("thirteen settle value err %s|%v", key, value)
				}
			}

			fmt.Printf("luaTest:%v|%v", key, value)

		})

		//test1 := string(test)
		//fmt.Printf("luaTest:%+v", test1)
	} else {
		fmt.Printf("luaTest:%+v:|%t", logic, ok)
	}
	// if err := L.DoString("return Logic:new()"); err != nil {
	// 	return err
	// }

	// if err := L.CallByParam(lua.P{
	// 	Fn:      L.GetGlobal("Logic:test1234"),
	// 	NRet:    1,
	// 	Protect: true,
	// }, lua.LNumber(10), lua.LNumber(11)); err != nil {
	// 	return err
	// }
	// ret := L.Get(-1) // returned value
	// L.Pop(1)         // remove received value
	// fmt.Printf("luaTest:%+v:", ret)
	return nil
}

func TestClean() error {
	cacher.FlushAll()
	// room.Status = enumr.RoomStatusDestroy
	f := func(tx *gorm.DB) error {
		dbr.DeleteAll(tx)
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return err
	}
	// err = cacheroom.DeleteRoom(pwd)
	// if err != nil {
	// 	return err
	// }

	// err = cachethirteen.DeleteGame(room.RoomID)
	// if err != nil {
	// 	return err
	// }
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
			RoomID:    room.RoomID,
			Status:    room.Status,
			Password:  room.Password,
			GameType:  room.GameType,
			CreatedAt: room.CreatedAt,
			List:      room.UserResults,
		}
		list = append(list, result.ToProto())
		//fmt.Printf("RoomResultListDDDDDD:%+v", result)
	}
	out := &pbr.RoomResultListReply{
		List: list,
	}
	return out, nil
}

func CheckRoomExist(uid int32) (*mdr.Room, error) {
	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func RoomUserStatusCheck() []*pbr.UserConnection {
	var ucs []*pbr.UserConnection
	rooms, err := GetRoomsByStatus(enumr.RoomStatusStarted)
	if err != nil {
		log.Err("room user status check err:%v", err)
		return nil
	}
	//fmt.Printf("RoomUserStatusCheck RoomNUmbers:%d", len(rooms))
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
				RoomID: room.RoomID,
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
	//fmt.Printf("RoomUserStatusCheck:%d", len(ucs))
	return ucs
}

func UpdateRoom(r *mdr.Room) (*mdr.Room, error) {
	f := func(tx *gorm.DB) error {
		room, err := dbr.UpdateRoom(tx, r)
		if err != nil {
			return err
		}
		r = room
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return r, err
}
