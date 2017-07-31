package thirteen

import (
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	enumt "playcards/model/thirteen/enum"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/db"

	"github.com/jinzhu/gorm"
)

func CreateThirteen(rid int32, userlist []*string) (*mdt.Thirteen, error) {
	var err error
	rooms := GetRoomsByStatusAndGameType()
	for _, rooms := range rooms {

	}
}

func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	f := func(tx *gorm.DB) error {
		list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
			enumr.RoomStatusStarted, enumt.ThirteenTableName)
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
