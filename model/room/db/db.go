package db

import (
	"playcards/model/room/enum"
	errr "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func CreateRoom(tx *gorm.DB, r *mdr.Room) error {
	if err := tx.Create(r).Error; err != nil {
		return errors.Internal("create room failed", err)
	}
	return nil
}

func UpdateRoom(tx *gorm.DB, r *mdr.Room) (*mdr.Room, error) {
	now := gorm.NowFunc()
	room := &mdr.Room{
		Users:     r.Users,
		Status:    r.Status,
		UpdatedAt: &now,
	}
	if err := tx.Model(r).Updates(room).Error; err != nil {
		return nil, errors.Internal("update room failed", err)
	}
	return r, nil
}

func GetRoomsByStatus(tx *gorm.DB, status int32) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetRoomsByStatusAndGameType(tx *gorm.DB, status int32,
	GameType int32) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where("status = ? and game_type = ?", status, GameType).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetRoomByID(tx *gorm.DB, rid int32) (*mdr.Room, error) {
	var (
		out mdr.Room
	)
	out.RoomID = rid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get room failed", err)
	}

	if !found {
		return nil, errr.ErrRoomNotExisted
	}
	return &out, nil
}

func BatchUpdate(tx *gorm.DB, status int32, ids []int32) error {
	sql, param, _ := squirrel.Update(enum.RoomTableName).
		Set("status", status).
		Where("room_id in (?)", ids).ToSql()
	err := tx.Exec(sql, param...).Error
	if err != nil {
		return errors.Internal("set room finish failed", err)
	}
	return nil
}
