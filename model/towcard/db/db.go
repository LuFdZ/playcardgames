package db

import (
	"playcards/model/towcard/enum"
	enumtow "playcards/model/towcard/enum"
	mdtow "playcards/model/towcard/mod"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateGame(tx *gorm.DB, game *mdtow.Towcard) error {
	now := gorm.NowFunc()
	game.UpdatedAt = &now
	if err := tx.Create(game).Error; err != nil {
		return errors.Internal("create tow card failed", err)
	}
	return nil
}

func UpdateGame(tx *gorm.DB, game *mdtow.Towcard) (*mdtow.Towcard, error) {
	now := gorm.NowFunc()
	towcard := &mdtow.Towcard{
		GameResultStr: game.GameResultStr,
		Status:     game.Status,
		UpdatedAt:  &now,
		OpDateAt:   game.OpDateAt,
	}
	if err := tx.Model(game).Updates(towcard).Error; err != nil {
		return nil, errors.Internal("update tow card failed", err)
	}
	return game, nil
}

func GetTowCardByRoomID(tx *gorm.DB, rid int32) ([]*mdtow.Towcard, error) {
	var out []*mdtow.Towcard
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get tow card by room_id failed", err)
	}
	return out, nil
}

func GetLastTowCardByRoomID(tx *gorm.DB, rid int32) (*mdtow.Towcard, error) {
	out := &mdtow.Towcard{}
	if err := tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
		return nil, errors.Internal("get last tow card by room_id failed", err)
	}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enumtow.TowCardTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get tow by room_id failed", err)
	}
	return nil
}

