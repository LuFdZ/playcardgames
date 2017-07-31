package db

import (
	errr "playcards/model/room/errors"
	"playcards/model/thirteen/enum"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func CreateThirteen(tx *gorm.DB, t *mdt.Thirteen) error {
	if err := tx.Create(t).Error; err != nil {
		return errors.Internal("create thirteen failed", err)
	}
	return nil
}

func UpdateThirteen(tx *gorm.DB, t *mdt.Thirteen) (*mdt.Thirteen, error) {
	now := gorm.NowFunc()
	thirteen := &mdt.Thirteen{
		UserList: t.UserList,
		Status:   t.Status,
		UpdateAt: &now,
	}
	if err := tx.Model(t).Updates(thirteen).Error; err != nil {
		return nil, errors.Internal("update thirteen failed", err)
	}
	return r, nil
}

func GetThirteensByStatus(tx *gorm.DB, status int32) ([]*mdt.Thirteen, error) {
	var (
		out []*mdt.Thirteen
	)
	if err := tx.Where("status = ?", status).Order("created_ed").
		Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetThitteenByID(tx *gorm.DB, tid int32) (*mdt.Thitreen, error) {
	var (
		out mdt.Thirteen
	)
	out.ThirteenID
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get thirteen failed", err)
	}

	if !found {
		return nil, errr.ErrRoomNotExisted
	}
	return &out, nil
}

func BatchUpdate(tx *gorm.DB, status int32, ids *[]int32) error {
	sql, param, _ := squirrel.Update(enum.ThirteenTableName).
		Set("status", status).
		Where("id in (?)", ids).ToSql()
	err := tx.Exec(sql, param...).Error
	if err != nil {
		return errors.Internal("set round finish failed", err)
	}
	return nil
}

func BatchCreate(tx *gorm.DB, status int32, ids *[]int32) error {
	sql, param, _ := squirrel.Update(enum.ThirteenTableName).
		Set("status", status).
		Where("id in (?)", ids).ToSql()
	err := tx.Exec(sql, param...).Error
	if err != nil {
		return errors.Internal("set round finish failed", err)
	}
	return nil
}
