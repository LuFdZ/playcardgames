package db
//
//import (
//	"github.com/jinzhu/gorm"
//	"playcards/utils/errors"
//	mdgr "playcards/model/goldroom/mod"
//)

//func CreateGRoom(tx *gorm.DB, r *mdgr.GoldRoom) error {
//	if err := tx.Create(r).Error; err != nil {
//		return errors.Internal("create gold room failed", err)
//	}
//	return nil
//}
//
//func UpdateGRoom(tx *gorm.DB, gr *mdgr.GoldRoom) (*mdgr.GoldRoom, error) {
//	now := gorm.NowFunc()
//	gr.CreatedAt = &now
//	gr.UpdatedAt = &now
//	gr := &mdgr.GoldRoom{
//		Users:     r.Users,
//		Status:    r.Status,
//		Giveup:    r.Giveup,
//		RoundNow:  r.RoundNow,
//		UpdatedAt: &now,
//		//CreatedAt: &now,
//		GiveupAt:  r.GiveupAt,
//		GameParam: r.GameParam,
//	}
//	if err := tx.Model(r).Updates(room).Error; err != nil {
//		return nil, errors.Internal("update room failed", err)
//	}
//	return r, nil
//}