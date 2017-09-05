package room

import (
	"fmt"
	"math/rand"
	dbbill "playcards/model/bill/db"
	cacheroom "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	dbt "playcards/model/thirteen/db"
	mdu "playcards/model/user/mod"
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

func CreateRoom(gtype int32, maxNum int32, roundNum int32, gparam string,
	user *mdu.User) (*mdr.Room,
	error) {
	var err error
	checkpwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
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
	pwd := GenerateRangeNum(100000, 999999)
	exist, err := cacheroom.CheckRoomExist(pwd)
	for i := 0; exist && i < 3; i++ {
		pwd = GenerateRangeNum(100000, 999999)
		exist, err = cacheroom.CheckRoomExist(pwd)
		if err != nil {
			return nil, err
		}
	}

	if exist {
		return nil, errors.ErrRoomPwdExisted
	}

	roomUser := GetRoomUser(user, enumr.UserUnready, 1,
		enumr.UserRoleMaster)
	users := []*mdr.RoomUser{roomUser}

	room := &mdr.Room{
		Password:    pwd,
		GameType:    gtype,
		MaxNumber:   maxNum,
		RoundNumber: roundNum,
		RoundNow:    1,
		GameParam:   gparam,
		//UserList:  fmt.Sprint(userID),
		Status: enumr.RoomStatusInit,
		Users:  users,
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
		cacheroom.DeleteRoom(room.Password)
		return nil, err
	}

	err = cacheroom.SetRoom(room)
	if err != nil {
		log.Err("room create set redis failed,%v | %v", room, err)
		return nil, err
	}

	cacheroom.SetRoomUser(room.RoomID, room.Password, user.UserID)
	//webroom.AutoSubscribe(u.UserID)
	return room, nil

}

func JoinRoom(pwd string, user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, nil, errors.ErrUserAlreadyInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, errors.ErrRoomNotExisted
	}

	//num := len(strings.Split(room.UserList, ","))
	num := len(room.Users)
	if num >= (int)(room.MaxNumber) {
		return nil, nil, errors.ErrRoomFull
	}

	//room.UserList = room.UserList + "," + fmt.Sprint(userID)
	position := 0
	isOrder := true
	for index, roomUser := range room.Users {
		if int32(index+1) != roomUser.Position {
			isOrder = false
			position = index + 1
			break
		}
	}
	if isOrder {
		position = num + 1
	}
	roomUser := GetRoomUser(user, enumr.UserUnready, int32(position),
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

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, nil, err
	}
	cacheroom.SetRoomUser(room.RoomID, room.Password, user.UserID)
	return roomUser, room, nil
}

func LeaveRoom(user *mdu.User) (*mdr.RoomUser, *mdr.Room, error) {
	pwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
	if len(pwd) == 0 {
		return nil, nil, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
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
		//若离开的是房主，解散房间
		if room.Users[i].UserID == user.UserID &&
			room.Status == enumr.RoomStatusInit {
			handle = 1
			roomUser = room.Users[i]
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
	cacheroom.DeleteRoomUser(room.RoomID, user.UserID)
	if isDestroy == 1 || len(newUsers) == 0 {
		room.Status = enumr.RoomStatusDestroy
		cacheroom.DeleteRoom(room.Password)
	} else {
		cacheroom.UpdateRoom(room)
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
	if err != nil {
		log.Err("room leave set session failed, %v", err)
		return nil, nil, err
	}

	return roomUser, room, nil
}

func GetReadyOrUnReady(pwd string, uid int32, readyStatus int32) (*mdr.
	RoomUser, int32, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(uid)
	//fmt.Printf("AAAAAAA: %d", checkpwd)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, 0, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, 0, err
	}
	if room.Status > enumr.RoomStatusInit {
		return nil, 0, errors.ErrNotReadyStatus
	}
	allReady := true
	t := time.Now()
	out := &mdr.RoomUser{}
	var num int32
	for _, user := range room.Users {
		num++
		if user.UserID == uid {
			user.Ready = readyStatus
			user.UpdatedAt = &t
			out = user

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

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, 0, err
	}
	return out, room.RoomID, nil
}

func GetRoomByUserID(uid int32, pwd string, InOrNotIn bool) (*mdr.Room, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(uid)
	if InOrNotIn {
		if len(checkpwd) == 0 {
			return nil, errors.ErrUserNotInRoom
		}

	} else {
		if len(pwd) != 0 {
			return nil, errors.ErrUserAlreadyInRoom
		}
	}

	room, err := cacheroom.GetRoom(pwd)
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
		err = cacheroom.UpdateRoom(room)
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
		//fmt.Printf("rooms round now:%d|%d", roundNow, room.RoundNumber)
		if room.RoundNow == room.RoundNumber {
			room.Status = enumr.RoomStatusDelay
			//fmt.Printf("rooms status now:%d", room.Status)
		} else {
			room.Status = enumr.RoomStatusInit
			room.RoundNow++
			for _, user := range room.Users {
				user.Ready = enumr.UserUnready
			}
		}

		f := func(tx *gorm.DB) error {
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
		if room.Status == enumr.RoomStatusDone {
			err = cacheroom.DeleteRoom(room.Password)
		} else {
			err = cacheroom.SetRoom(room)
		}

		if err != nil {
			log.Err("reinit update room redis err, %v", err)
			continue
		}
	}

	return rooms
}

func GiveUpGame(pwd string, status int32, uid int32) (*mdr.GiveUpGameResult,
	error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(uid)
	if len(checkpwd) == 0 || len(pwd) == 0 || pwd != checkpwd {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}

	if room.Status != enumr.RoomStatusStarted &&
		room.Status != enumr.RoomStatusWairtGiveUp {
		return nil, errors.ErrNotReadyStatus
	}
	//fmt.Printf("GiveUpGame:%d|%d", room.Status, enumr.RoomStatusStarted)
	giveUpResult := &room.GiveupGame
	if room.Status != enumr.RoomStatusWairtGiveUp {
		users := int32(len(room.Users))
		var userList []int32
		userList = append(userList, uid)
		giveUpResult := &mdr.GiveUpGameResult{
			RoomID:   room.RoomID,
			Total:    users,
			Agree:    1,
			DisAgree: 0,
			Wairting: users - 1,
			Status:   enumr.GiveupStatusWairting,
			Users:    userList,
		}
		room.GiveupGame = *giveUpResult
		room.Status = enumr.RoomStatusWairtGiveUp
	} else {
		for _, id := range room.GiveupGame.Users {
			if id == uid {
				return nil, errors.ErrAlreadyVoted
			}
		}
		giveUpResult.Users = append(giveUpResult.Users, uid)
		if status == enumr.GiveupStatusAgree {
			giveUpResult.Agree++
			giveUpResult.Wairting--
			if giveUpResult.Agree == giveUpResult.Total {
				room.Status = enumr.RoomStatusDestroy
				cacheroom.DeleteRoom(room.Password)
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
					return nil, err
				}
			}
		} else {
			room.Status = enumr.RoomStatusStarted
			giveUpResult.Status = enumr.GiveupStatusDisAgree
			giveUpResult.DisAgree++
			giveUpResult.Wairting--
			room.GiveupGame = *giveUpResult
		}
	}
	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}

	return giveUpResult, nil
}

func RoomDestroy() error {
	rooms, err := dbr.GetRoomsByStatus(db.DB(), enumr.RoomStatusReInit)
	if err != nil {
		log.Err("reinit get rooms by status err, %v", err)
		return nil
	}
	if len(rooms) == 0 {
		return nil
	}
	return nil
}

func RoomFinish(pwd string, status int32) error {
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return err
	}
	room.Status = status
	f := func(tx *gorm.DB) error {
		_, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return err
	}

	err = cacheroom.DeleteRoom(pwd)
	if err != nil {
		return err
	}
	return nil
}

func Heartbeat(uid int32) error {
	pwd := cacheroom.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
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

	err = cacheroom.UpdateRoom(room)
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
		Position:  position,
		Icon:      u.Icon,
		Sex:       u.Sex,
		Role:      role,
		UpdatedAt: &now,
	}
}

func Shock(uid int32, sendid int32) (int32, error) {
	pwd := cacheroom.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	pwdCheck := cacheroom.GetRoomPasswordByUserID(sendid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}

	if pwd != pwdCheck {
		return 0, errors.ErrNotInSameRoon
	}

	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return 0, err
	}
	if room.Status != enumr.RoomStatusStarted {
		return 0, errors.ErrNotReadyStatus
	}

	return room.RoomID, nil
}

func LuaTest() error {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoFile("lua/Logic.lua"); err != nil {
		return err
	}
	if err := L.DoString("return Logic:test1234(1,2)"); err != nil {
		return err
	}
	logic := L.Get(1)
	L.Pop(1)
	if test, ok := logic.(lua.LNumber); ok {
		test1 := int32(test)
		fmt.Printf("luaTest:%+v", test1)
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
	cacheroom.FlushAll()
	// room.Status = enumr.RoomStatusDestroy
	f := func(tx *gorm.DB) error {
		dbr.DeleteAll(tx)
		dbt.DeleteAll(tx)
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
