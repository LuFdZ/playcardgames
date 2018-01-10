package db

import (
	"playcards/model/fourcard/enum"
	enumfour "playcards/model/fourcard/enum"
	errn "playcards/model/fourcard/errors"
	mdfour "playcards/model/fourcard/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateGame(tx *gorm.DB, game *mdfour.FourCard) error {
	now := gorm.NowFunc()
	game.UpdatedAt = &now
	if err := tx.Create(game).Error; err != nil {
		return errors.Internal("create four card failed", err)
	}
	return nil
}

func UpdateGame(tx *gorm.DB, game *mdfour.FourCard) (*mdfour.FourCard, error) {
	now := gorm.NowFunc()
	fourcard := &mdfour.FourCard{
		GameResult: game.GameResult,
		Status:     game.Status,
		UpdatedAt:  &now,
		OpDateAt:   game.OpDateAt,
	}
	if err := tx.Model(game).Updates(fourcard).Error; err != nil {
		return nil, errors.Internal("update four card failed", err)
	}
	return game, nil
}

func GetFourCardByStatus(tx *gorm.DB, status int32) ([]*mdfour.FourCard, error) {
	var (
		out []*mdfour.FourCard
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get four card by status failed", err)
	}
	return out, nil
}

func GetFourCardAline(tx *gorm.DB) ([]*mdfour.FourCard, error) {
	var (
		out []*mdfour.FourCard
	)
	if err := tx.Where("status < ?", enumfour.GameStatusDone).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get four card by status failed", err)
	}
	return out, nil
}

func GetFourCardByID(tx *gorm.DB, gid int32) (*mdfour.FourCard, error) {
	var (
		out mdfour.FourCard
	)
	out.GameID = gid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get four card by id failed", err)
	}

	if !found {
		return nil, errn.ErrGameNotExist
	}
	return &out, nil
}

func GetFourCardByRoomID(tx *gorm.DB, rid int32) ([]*mdfour.FourCard, error) {
	var out []*mdfour.FourCard
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get four card by room_id failed", err)
	}
	return out, nil
}

func GetLastFourCardByRoomID(tx *gorm.DB, rid int32) (*mdfour.FourCard, error) {
	out := &mdfour.FourCard{}
	if err := tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
		return nil, errors.Internal("get last four card by room_id failed", err)
	}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enum.FourCardTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get niuniu by room_id failed", err)
	}
	return nil
}
