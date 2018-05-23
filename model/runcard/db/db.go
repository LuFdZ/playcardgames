package db

import (
	enumrun "playcards/model/runcard/enum"
	errun "playcards/model/runcard/errors"
	mdrun "playcards/model/runcard/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateGame(tx *gorm.DB, game *mdrun.Runcard) error {
	now := gorm.NowFunc()
	game.UpdatedAt = &now
	if err := tx.Create(game).Error; err != nil {
		return errors.Internal("create run card failed", err)
	}
	return nil
}

func UpdateGame(tx *gorm.DB, game *mdrun.Runcard) (*mdrun.Runcard, error) {
	now := gorm.NowFunc()
	fourcard := &mdrun.Runcard{
		GameResultStr: game.GameResultStr,
		Status:     game.Status,
		UpdatedAt:  &now,
		OpDateAt:   game.OpDateAt,
	}
	if err := tx.Model(game).Updates(fourcard).Error; err != nil {
		return nil, errors.Internal("update run card failed", err)
	}
	return game, nil
}

func GetRunCardByStatus(tx *gorm.DB, status int32) ([]*mdrun.Runcard, error) {
	var (
		out []*mdrun.Runcard
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get run card by status failed", err)
	}
	return out, nil
}

func GetRunCardAline(tx *gorm.DB) ([]*mdrun.Runcard, error) {
	var (
		out []*mdrun.Runcard
	)
	if err := tx.Where("status < ?", enumrun.GameStatusDone).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get run card by status failed", err)
	}
	return out, nil
}

func GetRunCardByID(tx *gorm.DB, gid int32) (*mdrun.Runcard, error) {
	var (
		out mdrun.Runcard
	)
	out.GameID = gid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get run card by id failed", err)
	}

	if !found {
		return nil, errun.ErrGameNotExist
	}
	return &out, nil
}

func GetRunCardByRoomID(tx *gorm.DB, rid int32) ([]*mdrun.Runcard, error) {
	var out []*mdrun.Runcard
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get run card by room_id failed", err)
	}
	return out, nil
}

func GetLastRunCardByRoomID(tx *gorm.DB, rid int32) (*mdrun.Runcard, error) {
	out := &mdrun.Runcard{}

	found, err := db.FoundRecord(tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get last run card by room_id failed", err)
	}

	if !found {
		return nil, nil
	}

	//if err := tx.Where(" room_id = ? ", rid).
	//	Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
	//	return nil, errors.Internal("get last four card by room_id failed", err)
	//}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enumrun.RunCardTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enumrun.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get run card by room_id failed", err)
	}
	return nil
}

