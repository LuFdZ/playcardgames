package room

import (
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
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
	err = CheckRoomByUserID(user.UserID)
	if err != nil {
		return nil, err
	}

	roomUser := GetRoomUser(user, 1, enumroom.UserUnready,
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

	f := func(tx *gorm.DB) error {
		err := dbr.CreateRoom(tx, room)
		if err != nil {
			return err
		}
		err = dbbill.GainBalance(tx, user.UserID,
			&mdbill.Balance{Gold: 20},
			enumbill.JournalTypeRoom, int64(user.UserID))
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, err
	}

	err = cacheroom.SetRoom(room)
	if err != nil {
		log.Err("room create set session failed, %v", err)
		return nil, err
	}
	return room, nil

}

func JoinRoom(pwd string, user *mdu.User) (*mdr.Room, error) {
	err := CheckRoomByUserID(user.UserID)
	if err != nil {
		return nil, err
	}

	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}

	//num := len(strings.Split(room.UserList, ","))
	num := len(room.Users)
	if num >= (int)(room.MaxNumber) {
		return nil, errors.ErrRoomFull
	}
	if num+1 == (int)(room.MaxNumber) {
		room.Status = enumroom.RoomStatusAllReady
	}

	//room.UserList = room.UserList + "," + fmt.Sprint(userID)
	position := -1
	for index, roomUser := range room.Users {
		if index > 0 && int32(index) != roomUser.Position {
			position = index
		}
	}
	roomUser := GetRoomUser(user, 1, int32(position), enumroom.UserRoleMaster)
	newUsers := append(room.Users, roomUser)
	room.Users = newUsers
	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		err = dbbill.GainBalance(tx, user.UserID,
			&mdbill.Balance{Gold: 20},
			enumbill.JournalTypeRoom, int64(user.UserID))

		room = r
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return nil, err
	}

	err = cacheroom.SetRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	return room, nil
}

func LeaveRoom(user *mdu.User) (*mdr.Room, error) {
	err := CheckRoomByUserID(user.UserID)
	if err != nil {
		return nil, err
	}

	room, err := GetRoomByUserID(user.UserID)
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
			} else {
				temp = newUsers
				newUsers = append(temp, room.Users[i])
			}
		}
	}
	if handle == 0 {
		return nil, errors.ErrNotInRoom
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
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	return room, nil
}

func GetReadyOrUnReady(pwd string, userID int32, readyStatus int32) (*mdr.RoomUser, error) {
	room, err := GetRoomByUserID(userID)
	if err != nil {
		return nil, err
	}
	if room.Status != enumroom.RoomStatusAllReady {
		return nil, errors.ErrNotReadyStatus
	}
	out := &mdr.RoomUser{}
	for _, user := range room.Users {
		if user.UserID == userID {
			user.Ready = readyStatus
			t := time.Now()
			user.UpdatedAt = &t
			out = user
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
		return nil, err
	}

	err = cacheroom.UpdateRoom(room)
	if err != nil {
		log.Err("room jooin set session failed, %v", err)
		return nil, err
	}
	return out, nil
}

func CheckRoomByUserID(userID int32) error {
	pwd := cacheroom.GetRoomPasswordByUserID(userID)
	if len(pwd) != 0 {
		return errors.ErrRoomAlreadyInRoom
	}
	return nil
}

func GetRoomByUserID(userID int32) (*mdr.Room, error) {
	pwd := cacheroom.GetRoomPasswordByUserID(userID)
	if len(pwd) == 0 {
		return nil, errors.ErrNotInRoom
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
	room, err := GetRoomByUserID(userID)
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
	return &mdr.RoomUser{
		UserID:   u.UserID,
		Ready:    ready,
		Position: position,
		Icon:     u.Icon,
		Sex:      u.Sex,
		Role:     role,
	}
}
