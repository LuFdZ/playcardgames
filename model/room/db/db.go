package db

import (
	mdr "playcards/model/room/mod"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateRoom(tx *gorm.DB, r *mdr.Room) error {
	if err := tx.Create(r).Error; err != nil {
		return errors.Internal("create room failed", err)
	}
	return nil
}

func UpdateRoom(tx *gorm.DB, r *mdr.Room) (*mdr.Room, error) {
	room := &mdr.Room{
		UserList: r.UserList,
		Status:   r.Status,
	}
	if err := tx.Model(r).Updates(room).Error; err != nil {
		return nil, errors.Internal("update room failed", err)
	}
	return r, nil
}
