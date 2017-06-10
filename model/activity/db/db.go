package db

import (
	erra "playcards/model/activity/errors"
	mda "playcards/model/activity/mod"
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
