package common

import (
	cachecon "playcards/model/common/cache"
	dbcon "playcards/model/common/db"
	enumcon "playcards/model/common/enum"
	errcon "playcards/model/common/errors"
	mdcon "playcards/model/common/mod"
	mdpage "playcards/model/page"
	"playcards/utils/db"

	"github.com/jinzhu/gorm"
)

func CreateBlackList(bltype int32, originid int32, targetid, uid int32) error {
	checkMbl, err := cachecon.GetBlackList(bltype, originid, targetid)
	if err != nil {
		return err
	}
	if checkMbl != nil {
		return errcon.ErrExisted
	}
	mbl := &mdcon.BlackList{
		Type:     bltype,
		OriginID: originid,
		TargetID: targetid,
		Status:   enumcon.ExamineStatusNew,
		OpID:     uid,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err := dbcon.CreateBlackList(tx, mbl)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = cachecon.SetBlackList(mbl)
	if err != nil {
		return err
	}
	return nil
}

func CancelBlackList(mbl *mdcon.BlackList, uid int32) error {
	checkMbl, err := cachecon.GetBlackList(mbl.Type, mbl.OriginID, mbl.TargetID)
	if err != nil {
		return err
	}
	if checkMbl == nil {
		return errcon.ErrNoExisted
	}
	mbl.OpID = uid
	mbl.Status = enumcon.BlackListStatusNoAvailable
	err = db.Transaction(func(tx *gorm.DB) error {
		mbl, err = dbcon.UpdateBlackList(tx, mbl)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = cachecon.DeleteBlackList(mbl)
	return nil
}

func CreateExamine(me *mdcon.Examine, uid int32) error {
	checkMe, err := cachecon.GetExamine(me.Type, me.AuditorID, me.ApplicantID)
	if err != nil {
		return err
	}
	if checkMe != nil {
		return errcon.ErrExisted
	}
	me.OpID = uid
	me.Status = enumcon.ExamineStatusNew
	err = db.Transaction(func(tx *gorm.DB) error {
		err := dbcon.CreateExamine(tx, me)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = cachecon.SetExamine(me)
	if err != nil {
		return err
	}
	return nil
}

func UpdateExamine(typeid int32, auditorid int32, applicantid, status int32, uid int32) error {
	me, err := cachecon.GetExamine(typeid, auditorid, applicantid)
	if err != nil {
		return err
	}
	if me == nil {
		return errcon.ErrNoExisted
	}
	me.OpID = uid
	me.Status = status
	if me.Status == enumcon.ExamineStatusNew {
		return errcon.ErrStatus
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		me, err = dbcon.UpdateExamine(tx, me)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = cachecon.DeleteExamine(me)
	return nil
}

func PageBlackList(page *mdpage.PageOption, mbl *mdcon.BlackList) (
	[]*mdcon.BlackList, int64, error) {
	return dbcon.PageBlackList(db.DB(), page, mbl)
}


func PageExamine(page *mdpage.PageOption, me *mdcon.Examine) (
	[]*mdcon.Examine, int64, error) {
	return dbcon.PageExamine(db.DB(), page, me)
}

func RefreshAllFromDB() error {
	mbls, err := dbcon.GetAllAlineBlackList(db.DB(), enumcon.TypeClub)
	if err != nil {
		return err
	}
	cachecon.SetAllBlackList(enumcon.TypeClub, mbls)
	mes, err := dbcon.GetAllAlineExamine(db.DB(), enumcon.TypeClub)
	if err != nil {
		return err
	}
	cachecon.SetAllExamine(enumcon.TypeClub, mes)
	return nil
}
