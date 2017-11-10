package db

import (
	mda "playcards/model/activity/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func GetPlayerShare(tx *gorm.DB, uid int32) (*mda.PlayerShare, error) {
	var out mda.PlayerShare
	found, err := db.FoundRecord(tx.Where("user_id = ?", uid).Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("player share get failed", err)
	}

	if !found {
		return nil, nil
	}

	return &out, nil
}

func CreatePlayerShare(tx *gorm.DB, uid int32) (*mda.PlayerShare, error) {
	ps := &mda.PlayerShare{
		UserID:        uid,
		ShareTimes:    0,
		TotalDiamonds: 0,
	}
	if err := tx.Create(ps).Error; err != nil {
		return nil, errors.Internal("player share failed", err)
	}
	return ps, nil
}

func UpdatePlayerShare(tx *gorm.DB, ps *mda.PlayerShare) error {
	now := gorm.NowFunc()
	ps.UpdatedAt = &now
	if err := tx.Model(ps).Updates(ps).Error; err != nil {
		return errors.Internal("update palyer share failed", err)
	}
	return nil
}
