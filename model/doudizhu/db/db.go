package db

import (
	mdddz "playcards/model/doudizhu/mod"
	"playcards/model/doudizhu/enum"
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


func GetLastDoudizhuByRoomID(tx *gorm.DB, rid int32) (*mdddz.Doudizhu, error) {
	out := &mdddz.Doudizhu{}
	if err := tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
		return nil, errors.Internal("get last doudizhu by room_id failed", err)
	}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enum.DoudizhuTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get doudizhu by room_id failed", err)
	}
	return nil
}

func GetDoudizhuByRoomID(tx *gorm.DB, rid int32) ([]*mdddz.Doudizhu, error) {
	var out []*mdddz.Doudizhu
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get doudizhu by room_id failed", err)
	}
	return out, nil
}