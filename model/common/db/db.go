package db

import (
	enumcommon "playcards/model/common/enum"
	mdCommon "playcards/model/common/mod"
	mdpage "playcards/model/page"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateBlackList(tx *gorm.DB, mbl *mdCommon.BlackList) error {
	now := gorm.NowFunc()
	mbl.CreatedAt = &now
	mbl.UpdatedAt = &now
	if err := tx.Create(mbl).Error; err != nil {
		return errors.Internal("create black list failed", err)
	}
	return nil
}

func UpdateBlackList(tx *gorm.DB, mbl *mdCommon.BlackList) (*mdCommon.BlackList, error) {
	now := gorm.NowFunc()
	mbl.UpdatedAt = &now
	if err := tx.Model(mbl).Updates(mbl).Error; err != nil {
		return nil, errors.Internal("update black list failed", err)
	}
	return mbl, nil
}

//func GetBlackList(tx *gorm.DB, me *mdCommon.Examine) (*mdCommon.Examine, error) {
//	var (
//		out mdCommon.Examine
//	)
//	found, err := db.FoundRecord(tx.Find(&out).
//		Where("applicant_id = ? and auditor_id = ? and status = ? and type = ?",
//		me.ApplicantID,me.AuditorID,me.Status,me.Type).Error)
//	if err != nil {
//		return nil, errors.Internal("get examine failed", err)
//	}
//
//	if !found {
//		return nil, nil
//	}
//	return &out, nil
//}

func PageBlackList(tx *gorm.DB, page *mdpage.PageOption,
	mbl *mdCommon.BlackList) ([]*mdCommon.BlackList, int64, error) {
	var out []*mdCommon.BlackList
	rows, rtx := page.Find(tx.Model(mbl).Order("created_at desc").
		Where(mbl), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page black list failed", rtx.Error)
	}
	return out, rows, nil
}

func CreateExamine(tx *gorm.DB, mde *mdCommon.Examine) error {
	now := gorm.NowFunc()
	mde.CreatedAt = &now
	mde.UpdatedAt = &now
	if err := tx.Create(mde).Error; err != nil {
		return errors.Internal("create examine failed", err)
	}
	return nil
}

func UpdateExamine(tx *gorm.DB, mde *mdCommon.Examine) (*mdCommon.Examine, error) {
	now := gorm.NowFunc()
	mde.UpdatedAt = &now
	if err := tx.Model(mde).Updates(mde).Error; err != nil {
		return nil, errors.Internal("update examine failed", err)
	}
	return mde, nil
}

//func GetExamine(tx *gorm.DB, me *mdCommon.Examine) (*mdCommon.Examine, error) {
//	var (
//		out mdCommon.Examine
//	)
//	found, err := db.FoundRecord(tx.Find(&out).
//	Where("applicant_id = ? and auditor_id = ? and status = ? and type = ?",
//		me.ApplicantID,me.AuditorID,me.Status,me.Type).Error)
//	if err != nil {
//		return nil, errors.Internal("get examine failed", err)
//	}
//
//	if !found {
//		return nil, nil
//	}
//	return &out, nil
//}

func PageExamine(tx *gorm.DB, page *mdpage.PageOption,
	mde *mdCommon.Examine) ([]*mdCommon.Examine, int64, error) {
	var out []*mdCommon.Examine
	rows, rtx := page.Find(tx.Model(mde).Order("created_at desc").
		Where(mde), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page examine failed", rtx.Error)
	}
	return out, rows, nil
}

func GetAllAlineBlackList(tx *gorm.DB, typeid int32) ([]*mdCommon.BlackList, error) {
	var (
		out []*mdCommon.BlackList
	)
	if err := tx.Where("type = ? and status = ?", typeid, enumcommon.BlackListStatusAvailable).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select black list failed", err)
	}
	return out, nil
}

func GetAllAlineExamine(tx *gorm.DB, typeid int32) ([]*mdCommon.Examine, error) {
	var (
		out []*mdCommon.Examine
	)
	if err := tx.Where("type = ? and status = ?", typeid, enumcommon.ExamineStatusNew).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select examine list failed", err)
	}
	return out, nil
}
