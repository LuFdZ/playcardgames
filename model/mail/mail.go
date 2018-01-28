package mail

import (
	dbgame "playcards/model/mail/db"
	mdgame "playcards/model/mail/mod"
	cachegame "playcards/model/mail/cache"
	enumgame "playcards/model/mail/enum"
	errgame "playcards/model/mail/errors"
	"playcards/utils/log"
	"playcards/utils/db"
	"github.com/jinzhu/gorm"
	"fmt"
)

func CreateMailInfo(mi *mdgame.MailInfo) (*mdgame.MailInfo, error) {
	return dbgame.CreateMailInfo(db.DB(), mi)
}

func UpdateMailInfo(mi *mdgame.MailInfo) (*mdgame.MailInfo, error) {
	if mi.MailID < 1 {
		return nil, errgame.ErrMailInfoNotFind
	}
	return dbgame.UpdateMailInfo(db.DB(), mi)
}

func SendMail(uid int32, msl *mdgame.MailSendLog, ids []int32, sendall int32) (*mdgame.MailSendLog, error) {
	if msl.MailType == enumgame.MailTypeSysMode && msl.MailID == 0 {
		return nil, errgame.ErrMailInfoID
	}
	if msl.MailType == enumgame.MailTypeSysMode {
		mi, err := cachegame.GetMailInfo(msl.MailID)
		if err != nil {
			return nil, err
		}
		if mi == nil {
			return nil, errgame.ErrMailInfoNotFind
		}
		msl.MailInfo = mi
	}
	if msl.MailInfo == nil {
		return nil, errgame.ErrMailInfoContent
	}
	var err error
	msl.SendID = uid
	msl.Status = enumgame.Success
	f := func(tx *gorm.DB) error {
		msl, err = dbgame.CreateMailSendLog(db.DB(), msl)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		log.Err("create mail send log failed, %v", err)
		return nil, nil
	}
	go createPlayerMail(msl, ids, sendall)
	return msl, err
}

func createPlayerMail(msl *mdgame.MailSendLog, ids []int32, sendall int32) {
	fmt.Printf("CreatePlayerMail:%v|%vids\n", msl, ids)
	var haveItem int32 = 0
	if len(msl.MailInfo.ItemList) > 0 {
		haveItem = 1
	}
	var pms []*mdgame.PlayerMail
	if sendall == 1 {
		ids = dbgame.GetAllUser(db.DB())
	}
	for _, id := range ids {
		pm := &mdgame.PlayerMail{
			SendLogID: msl.SendLogID,
			MailID:    msl.MailID,
			UserID:    id,
			SendID:    msl.SendID,
			MailType:  msl.MailType,
			Status:    enumgame.PlayermailNon,
			HaveItem:  haveItem,
		}
		pms = append(pms, pm)
	}
	dbgame.CreatePlayerMails(db.DB(), pms, msl)
}

func UpdateMailSendLog(msl *mdgame.MailSendLog) (*mdgame.MailSendLog, error) {
	if msl.SendLogID < 1 {
		return nil, errgame.ErrMailSendLogNotFind
	}
	return dbgame.UpdateMailSendLog(db.DB(), msl)
}

func RefreshAllMailInfoFromDB() error {
	mis, err := dbgame.GetMailInfos(db.DB())
	if err != nil {
		return err
	}

	return cachegame.SetMailInfos(mis)
}

func RefreshAllSendMailLogFromDB() error {
	msl, err := dbgame.GetMailSendLogs(db.DB())
	if err != nil {
		return err
	}
	return cachegame.SetMailSendLogs(msl)
}

func RefreshAllPlayerMailFromDB() error {
	pms, err := dbgame.GetPlayerMails(db.DB())
	if err != nil {
		return err
	}
	return cachegame.SetPlayerMails(pms)
}
