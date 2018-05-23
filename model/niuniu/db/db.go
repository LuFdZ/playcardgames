package db

import (
	"playcards/model/niuniu/enum"
	enumniu "playcards/model/niuniu/enum"
	errn "playcards/model/niuniu/errors"
	mdniu "playcards/model/niuniu/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateNiuniu(tx *gorm.DB, n *mdniu.Niuniu) error {
	now := gorm.NowFunc()
	n.UpdatedAt = &now
	if err := tx.Create(n).Error; err != nil {
		return errors.Internal("create niuniu failed", err)
	}
	return nil
}

func UpdateNiuniu(tx *gorm.DB, n *mdniu.Niuniu) (*mdniu.Niuniu, error) {
	now := gorm.NowFunc()
	niuniu := &mdniu.Niuniu{
		GameResults: n.GameResults,
		Status:      n.Status,
		UpdatedAt:   &now,
		OpDateAt:    n.OpDateAt,
	}
	if err := tx.Model(n).Updates(niuniu).Error; err != nil {
		return nil, errors.Internal("update niuniu failed", err)
	}
	return n, nil
}

func GetNiuniuByStatus(tx *gorm.DB, status int32) ([]*mdniu.Niuniu, error) {
	var (
		out []*mdniu.Niuniu
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get niuniu by status failed", err)
	}
	return out, nil
}

func GetNiuniuAline(tx *gorm.DB) ([]*mdniu.Niuniu, error) {
	var (
		out []*mdniu.Niuniu
	)
	if err := tx.Where("status < ?", enumniu.GameStatusDone).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("get niuniu by status failed", err)
	}
	return out, nil
}

func GetNiuniuByID(tx *gorm.DB, gid int32) (*mdniu.Niuniu, error) {
	var (
		out mdniu.Niuniu
	)
	out.GameID = gid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get niuniu by id failed", err)
	}

	if !found {
		return nil, errn.ErrGameNotExist
	}
	return &out, nil
}

func GetNiuniuByRoomID(tx *gorm.DB, rid int32) ([]*mdniu.Niuniu, error) {
	var out []*mdniu.Niuniu
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get niuniu by room_id failed", err)
	}
	return out, nil
}

func GetLastNiuniuByRoomID(tx *gorm.DB, rid int32) (*mdniu.Niuniu, error) {
	out := &mdniu.Niuniu{}

	found, err := db.FoundRecord(tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get last niuniu by room_id failed", err)
	}

	if !found {
		return nil, nil
	}

	//if err := tx.Where(" room_id = ? ", rid).
	//	Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
	//		//log.Err("GetLastNiuniuByRoomID fail rid:%d,err:%+v,date:%+v\n",rid,err,time.Now())
	//	return nil, errors.Internal("get last niuniu by room_id failed", err)
	//}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enum.NiuniuTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get niuniu by room_id failed", err)
	}
	return nil
}
