package db

import (
	erra "playcards/model/activity/errors"
	mda "playcards/model/activity/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func GetActivityConfig(tx *gorm.DB, id int32) (*mda.ActivityConfig, error) {
	var out mda.ActivityConfig
	if err := tx.Where("config_id = ?", id).Find(&out).Error; err != nil {
		return nil, errors.Internal("activity get failed", err)
	}
	return &out, nil
}

func AddActivityConfig(tx *gorm.DB, ac *mda.ActivityConfig) error {
	if err := tx.Create(ac).Error; err != nil {
		return errors.Internal("activity add failed", err)
	}

	return nil
}

func DeleteActivityConfig(tx *gorm.DB, ac *mda.ActivityConfig) error {
	dtx := tx.Delete(ac)
	if err := dtx.Error; err != nil {
		return errors.Internal("delete activity failed", err)
	}

	if rows := dtx.RowsAffected; rows == 0 {
		return erra.ErrActivityNotExisted
	}

	return nil
}

func UpdateActivityConfig(tx *gorm.DB, ac *mda.ActivityConfig) error {
	if err := tx.Model(ac).Updates(mda.ActivityConfig{
		Parameter:        ac.Parameter,
		LastModifyUserID: ac.LastModifyUserID,
	}).Error; err != nil {
		return errors.Internal("update activity failed", err)
	}
	return nil
}

func ActivityConfigList(tx *gorm.DB) ([]*mda.ActivityConfig, error) {
	var out []*mda.ActivityConfig
	if err := tx.Find(&out).Error; err != nil {
		return nil, errors.Internal("get activity list failed", err)
	}
	return out, nil
}

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
