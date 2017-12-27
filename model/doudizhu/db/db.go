package db

import (
	mdddz "playcards/model/doudizhu/mod"
	"playcards/utils/errors"
	"github.com/jinzhu/gorm"
)

func CreateDoudizhu(tx *gorm.DB, ddz *mdddz.Doudizhu) error {
	now := gorm.NowFunc()
	ddz.CreatedAt = &now
	ddz.UpdatedAt = &now
	if err := tx.Create(ddz).Error; err != nil {
		return errors.Internal("create doudizhu failed", err)
	}
	return nil
}

func UpdateDoudizhu(tx *gorm.DB, ddz *mdddz.Doudizhu) (*mdddz.Doudizhu, error) {
	now := gorm.NowFunc()
	doudizhu := &mdddz.Doudizhu{
		BankerID:    ddz.BankerID,
		Status:      ddz.Status,
		BankerTimes: ddz.BankerTimes,
		BombTimes:   ddz.BombTimes,
		OpID:        ddz.OpID,
		OpIndex:     ddz.OpIndex,
		WinerID:     ddz.WinerID,
		WinerType:   ddz.WinerType,
		OpDateAt:    ddz.OpDateAt,
		UpdatedAt:   &now,
	}
	if err := tx.Model(ddz).Updates(doudizhu).Error; err != nil {
		return nil, errors.Internal("update doudizhu failed", err)
	}
	return ddz, nil
}
