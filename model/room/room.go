package room

import (
	cacheroom "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	mdu "playcards/model/user/mod"
	"playcards/utils/db"
	"playcards/utils/log"
	"time"

	"github.com/jinzhu/gorm"
)

func CreateRoom(pwd string, gtype int32, num int32, user *mdu.User) (*mdr.Room,
	error) {
	var err error
	checkpwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 {
		return nil, errors.ErrUserAlreadyInRoom
	}

	roomUser := GetRoomUser(user, enumroom.UserUnready, 1,
		enumroom.UserRoleMaster)
	users := []*mdr.RoomUser{roomUser}

	room := &mdr.Room{
		Password:  pwd,
		GameType:  gtype,
		MaxNumber: num,
		//UserList:  fmt.Sprint(userID),
		Status: enumroom.RoomStatusInit,
		Users:  users,
	}

	exist, err := cacheroom.CheckRoomExist(pwd)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.ErrRoomPwdExisted
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

func JoinRoom(pwd string, user *mdu.User) (*mdr.Room, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
	if len(checkpwd) != 0 && pwd != checkpwd {
		return nil, errors.ErrUserAlreadyInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrRoomNotExisted
	}

	//num := len(strings.Split(room.UserList, ","))
	num := len(room.Users)
	if num >= (int)(room.MaxNumber) {
		return nil, errors.ErrRoomFull
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
	roomUser := GetRoomUser(user, enumroom.UserUnready, int32(position),
		enumroom.UserRoleSlave)
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
		return nil, err
	}

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	cacheroom.SetRoomUser(room.RoomID, room.Password, user.UserID)
	return room, nil
}

func LeaveRoom(user *mdu.User) (*mdr.Room, error) {
	pwd := cacheroom.GetRoomPasswordByUserID(user.UserID)
	if len(pwd) == 0 {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}

	if room.Status == enumroom.RoomStatusAllReady {
		return nil, errors.ErrGameHasBegin
	}

	newUsers := []*mdr.RoomUser{}
	temp := []*mdr.RoomUser{}
	handle := 0
	isDestroy := 0
	for i := range room.Users {
		//若离开的是房主，解散房间
		if room.Users[i].UserID == user.UserID {
			handle = 1
			if room.Users[i].Role == enumroom.UserRoleMaster {
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
		return nil, errors.ErrUserNotInRoom
	}
	if isDestroy == 1 || len(newUsers) == 0 {
		room.Status = enumroom.RoomStatusDestroy
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
		return nil, err
	}

	err = cacheroom.UpdateRoom(room)
	cacheroom.DeleteRoomUser(room.RoomID, user.UserID)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	return room, nil
}

func GetReadyOrUnReady(pwd string, userID int32, readyStatus int32) (*mdr.
	RoomUser, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(userID)
	if len(pwd) == 0 && pwd != checkpwd {
		return nil, errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	if room.Status > enumroom.RoomStatusInit {
		return nil, errors.ErrNotReadyStatus
	}
	allReady := true
	t := time.Now()
	out := &mdr.RoomUser{}
	var num int32
	for _, user := range room.Users {
		num++
		if user.UserID == userID {
			user.Ready = readyStatus
			user.UpdatedAt = &t
			out = user

		}
		if allReady && user.Ready == enumroom.UserUnready {
			allReady = false
		}
	}
	if allReady && num == room.MaxNumber {
		room.Status = enumroom.RoomStatusAllReady
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
		return nil, err
	}

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	return out, nil
}

func GetRoomByUserID(userID int32, password string, InOrNotIn bool) (*mdr.Room, error) {
	checkpwd := cacheroom.GetRoomPasswordByUserID(userID)
	if InOrNotIn {
		if len(checkpwd) == 0 {
			return nil, errors.ErrUserNotInRoom
		}

	} else {
		if len(password) != 0 {
			return nil, errors.ErrUserAlreadyInRoom
		}
	}

	room, err := cacheroom.GetRoom(password)
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

	f := func(tx *gorm.DB) error {
		rooms, err := dbr.GetRoomsByStatus(tx, enumroom.RoomStatusAllReady)
		if err != nil {
			return err
		}
		if len(rooms) == 0 {
			return nil
		}
		var ids []int32
		for _, room := range rooms {
			ids = append(ids, room.RoomID)
			room.Status = enumroom.RoomStatusStarted
			err = cacheroom.UpdateRoom(room)
			if err != nil {
				log.Err("room id:%d update status err:%v",
					room.RoomID, err)
				continue
			}
		}
		err = dbr.BatchUpdate(tx,
			enumroom.RoomStatusStarted, ids)

		if err != nil {
			log.Err("room update set session failed, %v", err)
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

func RoomStarted() {

}

func RoomDestroy() {

}

func RoomFinish(password string, status int32) error {
	room, err := cacheroom.GetRoom(password)
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

	err = cacheroom.DeleteRoom(password)
	if err != nil {
		return err
	}
	return nil
}

func Heartbrat(userID int32) error {
	pwd := cacheroom.GetRoomPasswordByUserID(userID)
	if len(pwd) == 0 {
		return errors.ErrUserNotInRoom
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return err
	}
	for _, item := range room.Users {
		if item.UserID == userID {
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
