package goldroom

import (
	"playcards/model/bill"
	enumbill "playcards/model/bill/enum"
	cacheroom "playcards/model/room/cache"
	cachegroom "playcards/model/goldroom/cache"
	cacheuser "playcards/model/user/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	enumgroom "playcards/model/goldroom/enum"
	//enumuser "playcards/model/user/enum"
	errroom "playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	mduser "playcards/model/user/mod"
	errorgame "playcards/model/goldroom/errors"
	gsync "playcards/utils/sync"
	mbill "playcards/model/bill/mod"
	pbroom "playcards/proto/room"

	"playcards/utils/db"
	"playcards/utils/log"

	"github.com/jinzhu/gorm"
	"time"
	"fmt"
	"playcards/utils/tools"
)

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func UpdateGoldRoom() map[*pbroom.RoomUser]int32 { //
	rooms := cachegroom.GetAllGRoom(0, []int32{enumroom.RoomStatusInit,
		enumroom.RoomStatusReInit})
	if rooms == nil && len(rooms) == 0 {
		return nil
	}
	m := make(map[*pbroom.RoomUser]int32)
	for _, mdr := range rooms {
		switch mdr.Status {
		case enumroom.RoomStatusInit:
			count := int32(len(mdr.ReadyUserMap))
			if count > 0 {
				numberNow := int32(len(mdr.Users))
				subTime := time.Now().Sub(*mdr.ReadyAt)

				if numberNow < enumgroom.GoldRoomInfoMap[mdr.GameType].MinNumber &&
					count == 1 {
					addTime := float64(tools.GenerateRangeNum(10, 30))
					if subTime.Seconds() < addTime {
						continue
					}
					robot, mdr, err := joinRobot(mdr.Password)
					if err != nil || robot == nil {
						log.Err("join robot err room:%v,err", mdr, err)
						continue
					}
					mdr.RobotIds = append(mdr.RobotIds, robot.UserID)
					msg := robot.ToProto()
					msg.Ids = mdr.PlayerIds
					m[msg] = enumgroom.UserOpJoin
					err = cacheroom.UpdateRoom(mdr)
					if err != nil {
						log.Err("room status init update room err:%v", err)
						continue
					}
				} else if count >= enumgroom.GoldRoomInfoMap[mdr.GameType].MinNumber {
					if subTime.Seconds() > 10 {
						mdr.Status = enumroom.RoomStatusAllReady
						f := func(tx *gorm.DB) error {
							_, err := dbroom.UpdateRoom(tx, mdr)
							if err != nil {
								log.Err("update gold room db err, %v|%v\n", err, mdr)
								return err
							}
							err = cacheroom.UpdateRoom(mdr)
							if err != nil {
								log.Err("room status init update room err:%v", err)
								return err
							}
							return nil
						}
						err := db.Transaction(f)
						if err != nil {
							log.Err("update gold room all ready transaction err, %v", err)
							continue
						}
					}
				}
			}

			for _, u := range mdr.Users {
				if (u.Type == 0 || u.Type == enumgroom.Player) && u.Ready == enumroom.UserUnready {
					stayTime := time.Now().Sub(*u.UpdatedAt)
					if stayTime.Seconds() > 60 {
						mdr, _ := removePlayer(mdr.Password, u.UserID)
						msg := u.ToProto()
						msg.Ids = mdr.PlayerIds
						m[msg] = enumgroom.UserOpRemove
					}
				}
				if u.Type == enumgroom.Robot && u.Ready == enumroom.UserUnready {
					stayTime := time.Now().Sub(*u.UpdatedAt)
					addTime := float64(tools.GenerateRangeNum(5, 10))
					if stayTime.Seconds() > addTime {
						robot, err := robotSetSit(u.UserID, mdr.Password, enumroom.RoomStatusAllReady)
						if err != nil {
							log.Err("gold room set robot ready fail robot:%+v,err:%+v\n", robot, err)
							continue
						}
						msg := robot.ToProto()
						msg.Ids = mdr.PlayerIds
						m[msg] = enumgroom.UserOpReady
					}
				}
			}
			break
		case enumroom.RoomStatusReInit:
			f := func(tx *gorm.DB) error {
				//更新玩家游戏局数
				err := dbroom.UpdateRoomPlayTimes(tx, mdr.RoomID, mdr.GameType)
				if err != nil {
					log.Err("update gold room play times db err, %v|%v\n", err)
					return err
				}
				t := time.Now()
				for _, u := range mdr.Users {

					if u.Type == enumgroom.Player {
						err := bill.GainGameBalance(u.UserID, mdr.RoomID, mdr.GameType*100+1,
							&mbill.Balance{Amount: int64(u.ResultAmount), CoinType: enumbill.TypeGold})
						if err != nil {
							return err
						}
					} else {
						mdr, _ = removePlayer(mdr.Password, u.UserID)
						msg := u.ToProto()
						msg.Ids = mdr.PlayerIds
						m[msg] = enumgroom.UserOpRemove
					}
					if _, ok := mdr.ReadyUserMap[u.UserID]; ok {
						u.Ready = enumroom.UserUnready
					}

					u.UpdatedAt = &t
				}

				mdr.Status = enumroom.RoomStatusInit
				//mdr.UpdatedAt = &t
				_, err = dbroom.UpdateRoom(tx, mdr)
				if err != nil {
					log.Err("update gold room db err, %v|%v\n", err, mdr)
					return err
				}
				fmt.Printf("AAAAAAAAAAAAAAARoomStatusReInit:%+v\n", mdr)
				err = cacheroom.UpdateRoom(mdr)
				if err != nil {
					log.Err("room status init update room err:%v", err)
					return err
				}
				return nil
			}
			err := db.Transaction(f)
			if err != nil {
				log.Err("update gold room transaction err, %v", err)
				continue
			}
			break
		}
	}
	return m
}

func JoinRoom(gtype int32, level int32, mdu *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	hasRoom := cacheroom.ExistRoomUser(mdu.UserID)
	if hasRoom {
		return nil, nil, errroom.ErrUserAlreadyInRoom
	}
	paramRight := checkRoomParam(gtype, level)
	if !paramRight {
		return nil, nil, errroom.ErrGameParam
	}

	limit := enumgroom.GoldRoomCostMap[gtype][level][1]
	userBalance, err := bill.GetUserBalance(mdu.UserID, enumbill.TypeGold)
	if err != nil {
		return nil, nil, err
	}
	if userBalance.Balance < limit {
		return nil, nil, errroom.ErrNotEnoughGold
	}
	mdr, err := cachegroom.SelectGRoom(gtype, level)
	if err != nil {
		return nil, nil, err
	}

	if mdr == nil {
		mdr, err = CreateGoldRoom(gtype, level)
		if err != nil {
			return nil, nil, err
		}
	}
	mdr.WatchIds = append(mdr.WatchIds, mdu.UserID)
	//mdru, mdr, err := roomAddUserAndSeating(mdu, mdr.Password)
	//if err != nil {
	//	return nil, nil, err
	//}
	//mdru.Gold = userBalance.Balance
	return roomAddUserAndSeating(mdu, mdr.Password) //mdru, mdr, nil
}

func SetReady(uid int32, pwd string) ([]int32, error) {
	mdr, err := cacheroom.GetRoomUserID(uid)
	if err != nil {
		return nil, errroom.ErrUserNotInRoom
	}
	if pwd != mdr.Password {
		return nil, errroom.ErrUserNotInRoom
	}

	var ids []int32
	f := func() error {
		mdr, err := cacheroom.GetRoom(pwd)
		if err != nil {
			return err
		}
		allReady := true
		t := time.Now()
		var num int32
		//mdr.WatchIds = append(mdr.WatchIds,mdu.UserID)
		var watchIds []int32
		for _, user := range mdr.Users {
			num++
			if user.UserID == uid {
				if user.Ready == enumroom.UserReady {
					return errroom.ErrAlreadyReady
				}
				user.Ready = enumroom.UserReady
				user.UpdatedAt = &t
				user.Join = enumgroom.JoinGame
				if mdr.Status > enumroom.RoomStatusInit {
					user.Join = enumgroom.UnJoinGame
				}
			} else if user.Ready == enumroom.UserUnready {
				watchIds = append(watchIds, user.UserID)
			}
			if allReady && user.Ready != enumroom.UserReady {
				allReady = false
			}
		}
		if allReady && num == mdr.MaxNumber {
			mdr.Status = enumroom.RoomStatusAllReady
		}
		mdr.ReadyUserMap[uid] = t.Unix()
		if len(mdr.ReadyUserMap) <= 2 {
			mdr.ReadyAt = &t
		}
		ids = mdr.Ids
		mdr.WatchIds = watchIds
		err = cacheroom.UpdateRoom(mdr)
		if err != nil {
			log.Err("update set sit failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
			return err
		}
		return nil
	}
	lock := RoomLockKey(mdr.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		return ids, err
	}
	return ids, nil
}

func robotSetSit(uid int32, pwd string, roomStatus int32) (*mdroom.RoomUser, error) {

	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if pwd != mdr.Password {
		return nil, errroom.ErrUserNotInRoom
	}

	robotInRoom := false
	for _, robot := range mdr.Users {
		if robot.UserID == uid {
			robotInRoom = true
		}
	}
	if !robotInRoom {
		return nil, errroom.ErrUserNotInRoom
	}

	if mdr.Status > enumroom.RoomStatusInit {
		return nil, errroom.ErrNotReadyStatus
	}

	roomUser := &mdroom.RoomUser{}
	f := func() error {
		t := time.Now()
		var num int32
		mdrLock, err := cacheroom.GetRoom(pwd)
		if err != nil {
			return err
		}
		for _, u := range mdrLock.Users {
			num++
			if u.UserID == uid {
				if u.Ready == enumroom.UserReady {
					return errroom.ErrAlreadyReady
				}
				u.Ready = enumroom.UserReady
				u.UpdatedAt = &t
				roomUser = u
			}
		}
		mdrLock.Status = roomStatus
		err = cacheroom.UpdateRoom(mdrLock)
		if err != nil {
			log.Err("set robot room ready failed, roomid:%d,uid:%d,err:%v\n", mdr.RoomID, uid, err)
			return err
		}
		mdr = mdrLock
		return nil
	}
	lock := RoomLockKey(pwd)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return nil, err
	}
	return roomUser, nil
}

func joinRobot(pwd string) (*mdroom.RoomUser, *mdroom.Room, error) {
	mdr, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, nil, err
	}
	robot, err := cacheuser.GetRobot()
	if err != nil {
		return nil, nil, err
	}
	if robot == nil {
		return nil, nil, errorgame.ErrRobotNotFind
	}
	return roomAddUserAndSeating(robot, mdr.Password)
}

func roomAddUserAndSeating(mdu *mduser.User, pwd string) (*mdroom.RoomUser, *mdroom.Room, error) {
	var roomUser *mdroom.RoomUser
	var mdrReturn *mdroom.Room
	f := func() error {
		mdr, err := cacheroom.GetRoom(pwd)
		if err != nil {
			return err
		}
		positionArray := make([]int32, mdr.MaxNumber)
		for n := 0; n < len(positionArray); n++ {
			positionArray[n] = 0
		}
		for _, ru := range mdr.Users {
			if ru.UserID == mdu.UserID {
				return errroom.ErrUserAlreadyInRoom
			}
			positionArray[ru.Position-1] = 1
		}
		p := 0
		for n := 0; n < len(positionArray); n++ {
			if positionArray[n] == 0 {
				p = n
			}
		}
		//ready := enumroom.UserUnready
		//if mdu.Type == enumuser.Robot {
		//	ready = enumroom.UserReady
		//}
		ru := GetRoomUser(mdu, enumroom.UserUnready, int32(p+1),
			enumroom.UserRoleSlave, mdu.Type)
		if ru.Type == enumgroom.Robot {
			ru.Nickname = mdu.Nickname
			ru.Sex = mdu.Sex
		} else {
			mdr.PlayerIds = append(mdr.PlayerIds, ru.UserID)
		}
		roomUser = ru
		mdr.Users = append(mdr.Users, roomUser)
		mdr.Ids = append(mdr.Ids, mdu.UserID)
		if mdr.RoomType != enumroom.RoomTypeNom && len(mdr.Users) == 0 {
			roomUser.Role = enumroom.UserRoleMaster
		}
		err = cacheroom.UpdateRoom(mdr)
		if err != nil {
			log.Err("room join set session failed, %v|%v\n", err, mdr)
			return err
		}
		//if mdu.Type == enumuser.Player {
		//	err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mdu.UserID)
		//	if err != nil {
		//		log.Err("room user join set session failed, %v|%v\n", err, mdr)
		//		return err
		//	}
		//}
		err = cacheroom.SetRoomUser(mdr.RoomID, mdr.Password, mdu.UserID)
		if err != nil {
			log.Err("room user join set session failed, %v|%v\n", err, mdr)
			return err
		}
		//UpdateRoom(mdr)
		mdrReturn = mdr
		return nil
	}
	lock := RoomLockKey(pwd)
	err := gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return nil, nil, err
	}
	return roomUser, mdrReturn, nil
}

func LeaveRoom(mduser *mduser.User) (*mdroom.RoomUser, *mdroom.Room, error) {
	mdrReturn, err := cacheroom.GetRoomUserID(mduser.UserID)
	if err != nil {
		return nil, nil, err
	}
	if mdrReturn == nil {
		return nil, nil, errroom.ErrUserNotInRoom
	}
	if mdrReturn.Status > enumroom.RoomStatusAllReady && mdrReturn.Status < enumroom.RoomStatusReInit {
		return nil, nil, errroom.ErrGameHasBegin
	}
	roomUser := &mdroom.RoomUser{}

	f := func() error {
		newUsers := []*mdroom.RoomUser{}
		ids := []int32{}
		mdr, _ := cacheroom.GetRoom(mdrReturn.Password)
		for _, u := range mdr.Users {
			if u.UserID != mduser.UserID {
				newUsers = append(newUsers, u)
				ids = append(ids, u.UserID)
			} else {
				roomUser = u
			}
		}
		mdr.Users = newUsers
		mdr.Ids = ids
		for _, user := range mdr.Users {
			mdr.Ids = append(mdr.Ids, user.UserID)
			if user.Type == enumgroom.Player {
				mdr.PlayerIds = append(mdr.PlayerIds, user.UserID)
			}
		}

		if len(mdr.PlayerIds) == 0 {
			mdr.Users = nil
			mdr.Status = enumroom.RoomStatusDestroy
			err := cacheroom.DeleteAllRoomUser(mdr.Password, "LeaveRoom")
			if err != nil {
				log.Err("leave room delete all users failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
			err = cacheroom.DeleteRoom(mdr)
			if err != nil {
				log.Err("leave room delete room failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
		} else {
			err = cacheroom.DeleteRoomUser(mduser.UserID)
			if err != nil {
				log.Err("room leave room delete user failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
			if mdr.RoomType != enumroom.RoomTypeNom && len(newUsers) > 0 {
				newUsers[0].Role = enumroom.UserRoleMaster
			}
			err := cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("room leave room delete failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}

		}
		updateRoom(mdr)
		mdrReturn = mdr
		return nil
	}
	lock := RoomLockKey(mdrReturn.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return nil, nil, err
	}
	return roomUser, mdrReturn, nil
}

func removePlayer(pwd string, uid int32) (*mdroom.Room, error) {
	mdrReturn, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	//roomUser := &mdroom.RoomUser{}
	f := func() error {
		newUsers := []*mdroom.RoomUser{}
		mdr, _ := cacheroom.GetRoom(mdrReturn.Password)
		//var ids []int32
		for _, u := range mdr.Users {
			if u.UserID != uid {
				newUsers = append(newUsers, u)
				//ids = append(ids,u.UserID)
			}
		}
		mdr.Users = newUsers
		mdr.Ids = []int32{}
		mdr.PlayerIds = []int32{}
		mdr.RobotIds = []int32{}
		for _, mdu := range mdr.Users {
			if mdu.UserID != uid {
				mdr.Ids = append(mdr.Ids, mdu.UserID)
				if mdu.Type == enumgroom.Player {
					mdr.PlayerIds = append(mdr.PlayerIds, mdu.UserID)
				} else if mdu.Type == enumgroom.Robot {
					mdr.RobotIds = append(mdr.RobotIds, mdu.UserID)
				}
			}
		}
		newUserResult := []*mdroom.GameUserResult{}
		for _, ur := range mdr.UserResults {
			if ur.UserID != uid {
				newUserResult = append(newUserResult, ur)
			}
		}
		mdr.UserResults = newUserResult
		if len(mdr.PlayerIds) == 0 {
			mdr.Users = nil
			mdr.Status = enumroom.RoomStatusDestroy
			err := cacheroom.DeleteAllRoomUser(mdr.Password, "LeaveRoom")
			if err != nil {
				log.Err("leave room delete all users failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
			err = cacheroom.DeleteRoom(mdr)
			if err != nil {
				log.Err("leave room delete room failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
		} else {
			err = cacheroom.DeleteRoomUser(uid)
			if err != nil {
				log.Err("room leave room delete user failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
			if mdr.RoomType != enumroom.RoomTypeNom && len(newUsers) > 0 {
				newUsers[0].Role = enumroom.UserRoleMaster
			}
			err := cacheroom.UpdateRoom(mdr)
			if err != nil {
				log.Err("room leave room delete failed, %d|%v\n",
					mdr.RoomID, err)
				return err
			}
		}
		updateRoom(mdr)
		mdrReturn = mdr
		return nil
	}
	lock := RoomLockKey(mdrReturn.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return nil, err
	}
	return mdrReturn, err
}

func checkRoomParam(gtype int32, level int32) bool {
	if gtype != 1001 && gtype != 1002 && gtype != 1003 && gtype != 1004 {
		return false
	}
	if level != 1 && level != 2 {
		return false
	}
	return true
}

func GetRoomUser(mdu *mduser.User, ready int32, position int32,
	role int32, typ int32) *mdroom.RoomUser {
	t := time.Now()
	ru := &mdroom.RoomUser{
		UserID:    mdu.UserID,
		Ready:     ready,
		Position:  position,
		Role:      role,
		Type:      typ,
		UpdatedAt: &t,
	}
	return ru
}

func updateRoom(room *mdroom.Room) error {
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

func CreateGoldRoom(gtype int32, level int32) (*mdroom.Room, error) {
	users := []*mdroom.RoomUser{}
	ids := []int32{}
	gameInfo := enumgroom.GoldRoomInfoMap[gtype]
	t := time.Now()
	pwd, err := getPassWord()
	if err != nil {
		return nil, err
	}

	mdr := &mdroom.Room{
		Password:       pwd,
		GameType:       gtype,
		MaxNumber:      gameInfo.MaxNumber,
		RoundNumber:    0,
		RoundNow:       1,
		GameParam:      gameInfo.RoomParam,
		Status:         enumroom.RoomStatusInit,
		Giveup:         enumroom.NoGiveUp,
		Users:          users,
		RoomType:       enumroom.RoomTypeGold,
		GiveupAt:       &t,
		Ids:            ids,
		Cost:           0,
		Level:          level,
		StartMaxNumber: gameInfo.MaxNumber,
		CostType:       enumroom.CostTypeGold,
		Flag:           enumroom.RoomNoFlag,
		ReadyUserMap:   make(map[int32]int64),
	}
	f := func(tx *gorm.DB) error {
		err := dbroom.CreateRoom(tx, mdr)
		if err != nil {
			log.Err("room create failed,%v | %v", mdr, err)
			return err
		}
		err = cacheroom.SetRoom(mdr)
		if err != nil {
			log.Err("room create set redis failed,%v | %v\n", mdr, err)
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return mdr, nil
}

func getPassWord() (string, error) {
	pwdNew := fmt.Sprintf("%d", tools.GenerateRangeNum(enumroom.RoomCodeMin, enumroom.RoomCodeMax))
	exist := cacheroom.CheckRoomExist(pwdNew)
	for i := 0; exist && i < 3; i++ {
		pwdNew = fmt.Sprintf("%d", tools.GenerateRangeNum(enumroom.RoomCodeMin, enumroom.RoomCodeMax))
		exist = cacheroom.CheckRoomExist(pwdNew)
	}
	if exist {
		return "", errroom.ErrRoomPwdExisted
	}
	return pwdNew, nil
}
