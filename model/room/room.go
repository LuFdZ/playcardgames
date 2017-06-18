package room

import (
	"fmt"
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacheroom "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	mdr "playcards/model/room/mod"
	"playcards/utils/db"
	"playcards/utils/errors"
	"playcards/utils/log"
	"strings"

	"github.com/jinzhu/gorm"
)

func CreateRoom(pwd string, gtype int32, num int32, userID int32) (*mdr.Room,
	error) {
	var err error
	room := &mdr.Room{
		Password:     pwd,
		GameType:     gtype,
		PlayerMaxNum: num,
		UserList:     fmt.Sprint(userID),
		Status:       enumroom.Waiting,
	}
	f := func(tx *gorm.DB) error {
		err := dbr.CreateRoom(tx, room)
		if err != nil {
			return err
		}
		err = dbbill.GainBalance(tx, userID,
			&mdbill.Balance{Gold: 20},
			enumbill.JournalTypeRoom, int64(userID))
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

func JoinRoom(pwd string, userID int32) (*mdr.Room, error) {
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return nil, err
	}
	num := len(strings.Split(room.UserList, ","))
	if num >= (int)(room.PlayerMaxNum) {
		return nil, errors.Internal("room full", err)
	}
	room.UserList = room.UserList + "," + fmt.Sprint(userID)
	f := func(tx *gorm.DB) error {
		r, err := dbr.UpdateRoom(tx, room)
		if err != nil {
			return err
		}
		err = dbbill.GainBalance(tx, userID,
			&mdbill.Balance{Gold: 20},
			enumbill.JournalTypeRoom, int64(userID))

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

func RoomEnd(pwd string) error {
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return err
	}
	room.Status = enumroom.End
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
	return nil
}
