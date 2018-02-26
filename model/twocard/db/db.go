package db

import (
	"playcards/model/twocard/enum"
	enumtow "playcards/model/twocard/enum"
	mdtwo "playcards/model/twocard/mod"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateGame(tx *gorm.DB, game *mdtwo.Twocard) error {
	now := gorm.NowFunc()
	game.UpdatedAt = &now
	if err := tx.Create(game).Error; err != nil {
		return errors.Internal("create two card failed", err)
	}
	return nil
}

func UpdateGame(tx *gorm.DB, game *mdtwo.Twocard) (*mdtwo.Twocard, error) {
	now := gorm.NowFunc()
	towcard := &mdtwo.Twocard{
		GameResultStr: game.GameResultStr,
		Status:     game.Status,
		UpdatedAt:  &now,
		OpDateAt:   game.OpDateAt,
	}
	if err := tx.Model(game).Updates(towcard).Error; err != nil {
		return nil, errors.Internal("update two card failed", err)
	}
	return game, nil
}

func GetTwoCardByRoomID(tx *gorm.DB, rid int32) ([]*mdtwo.Twocard, error) {
	var out []*mdtwo.Twocard
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get two card by room_id failed", err)
	}
	return out, nil
}

func GetLastTwoCardByRoomID(tx *gorm.DB, rid int32) (*mdtwo.Twocard, error) {
	out := &mdtwo.Twocard{}
	if err := tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
		return nil, errors.Internal("get last two card by room_id failed", err)
	}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enumtow.TowCardTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get two by room_id failed", err)
	}
	return nil
}

