package db

import (
	mdmail "playcards/model/mail/mod"
	enummail "playcards/model/mail/enum"
	mdpage "playcards/model/page"
	cachemail"playcards/model/mail/cache"
	"github.com/jinzhu/gorm"
	"playcards/utils/errors"
	"playcards/utils/log"
	"github.com/Masterminds/squirrel"
	"fmt"
)

func CreateMailInfo(tx *gorm.DB, mdmi *mdmail.MailInfo) (*mdmail.MailInfo, error) {
	now := gorm.NowFunc()
	mdmi.Status = enummail.MailInfoStatusNom
	mdmi.UpdatedAt = &now
	mdmi.CreatedAt = &now
	if err := tx.Create(mdmi).Error; err != nil {
		return nil, errors.Internal("create mail info failed", err)
	}
	return mdmi, nil
}

func UpdateMailInfo(tx *gorm.DB, mdmi *mdmail.MailInfo) (*mdmail.MailInfo, error) {
	now := gorm.NowFunc()
	mi := &mdmail.MailInfo{
		MailName:    mdmi.MailName,
		MailTitle:   mdmi.MailTitle,
		MailContent: mdmi.MailContent,
		MailType:    mdmi.MailType,
		ItemList:    mdmi.ItemList,
		UpdatedAt:   &now,
	}
	if err := tx.Model(mdmi).Updates(mi).Error; err != nil {
		return nil, errors.Internal("update mail info failed", err)
	}
	return mi, nil
}

func CreateMailSendLog(tx *gorm.DB, mdmsl *mdmail.MailSendLog) (*mdmail.MailSendLog, error) {
	now := gorm.NowFunc()
	mdmsl.Status = enummail.MailSendNom
	mdmsl.UpdatedAt = &now
	mdmsl.CreatedAt = &now
	if err := tx.Create(mdmsl).Error; err != nil {
		return nil, errors.Internal("create mail send log", err)
	}
	return mdmsl, nil
}

func UpdateMailSendLog(tx *gorm.DB, mdmsl *mdmail.MailSendLog) (*mdmail.MailSendLog, error) {
	now := gorm.NowFunc()
	msl := &mdmail.MailSendLog{
		SendCount:  mdmsl.SendCount,
		TotalCount: mdmsl.TotalCount,
		UpdatedAt:  &now,
	}
	if err := tx.Model(mdmsl).Updates(msl).Error; err != nil {
		return nil, errors.Internal("update mail send log failed", err)
	}
	return mdmsl, nil
}

func GetAllUser(tx *gorm.DB,channel string) []int32 {
	var (
		ids []int32
	)
	sql, param, err := squirrel.
	Select(" user_id").From("users").Where("channel = ?",channel).ToSql()
	if err != nil {
		log.Err("install player mails get user id squirrel failed")
	}

	err = tx.Raw(sql, param...).Scan(&ids).Error
	if err != nil {
		log.Err("install player mails get user id list failed")
	}
	return ids
}

func CreatePlayerMails(tx *gorm.DB, mdpms []*mdmail.PlayerMail, mdmsl *mdmail.MailSendLog) {
	fmt.Sprintf("AAAAACreatePlayerMails\n")
	now := gorm.NowFunc()
	var count int32 = 0
	for _, mdpm := range mdpms {
		mdpm.UpdatedAt = &now
		mdpm.CreatedAt = &now
		if err := tx.Create(mdpm).Error; err != nil {
			return
			log.Err("create player mail failed! uid:%d,send_log_id:%d,err:%+v", mdpm.UserID, mdpm.SendLogID, err)
		}
		fmt.Printf("AAAACreatePlayerMails\n")
		err := cachemail.SetPlayerMail(mdpm)
		if err != nil {
			return
			log.Err("create player mail redis failed! uid:%d,send_log_id:%d,err:%+v", mdpm.UserID, mdpm.SendLogID, err)
		}
		fmt.Printf("BBBBCreatePlayerMails\n")
		count ++
	}
	mdmsl.SendCount = count
	mdmsl.TotalCount = int32(len(mdpms))
	_, _ = UpdateMailSendLog(tx, mdmsl)
	err := cachemail.SetMailSendLog(mdmsl)
	if err != nil {
		log.Err("create mail send log redis failed! send log ID:%d,err:%+v", mdmsl.LogID, err)
		return
	}
}

func UpdatePlayerMail(tx *gorm.DB, mdpm *mdmail.PlayerMail) (*mdmail.PlayerMail, error) {
	now := gorm.NowFunc()
	pm := &mdmail.PlayerMail{
		Status:    mdpm.Status,
		UpdatedAt: &now,
	}
	if err := tx.Model(mdpm).Updates(pm).Error; err != nil {
		return nil, errors.Internal("update player mail failed", err)
	}
	return pm, nil
}

func PageMailInfos(tx *gorm.DB, page *mdpage.PageOption,
	n *mdmail.MailInfo) ([]*mdmail.MailInfo, int64, error) {
	var out []*mdmail.MailInfo
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page mail info failed", rtx.Error)
	}
	return out, rows, nil
}

func PageMailSendLogs(tx *gorm.DB, page *mdpage.PageOption,
	n *mdmail.MailSendLog) ([]*mdmail.MailSendLog, int64, error) {
	var out []*mdmail.MailSendLog
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page send mail log failed", rtx.Error)
	}
	return out, rows, nil
}

func PagePlayerMails(tx *gorm.DB, page *mdpage.PageOption,
	n *mdmail.PlayerMail) ([]*mdmail.PlayerMail, int64, error) {
	var out []*mdmail.PlayerMail
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page player mail log failed", rtx.Error)
	}
	return out, rows, nil
}


func GetMailInfos(tx *gorm.DB) ([]*mdmail.MailInfo, error) {
	var (
		out []*mdmail.MailInfo
	)
	if err := tx.Where("status = ?", enummail.MailTypeSysMode).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select mail info list failed", err)
	}
	return out, nil
}

func GetMailSendLogs(tx *gorm.DB) ([]*mdmail.MailSendLog, error) {
	var (
		out []*mdmail.MailSendLog
	)
	if err := tx.Where("status = ?", enummail.MailTypeSysMode).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select send mail log list failed", err)
	}
	return out, nil
}

func GetPlayerMails(tx *gorm.DB) ([]*mdmail.PlayerMail, error) {
	var (
		out []*mdmail.PlayerMail
	)
	if err := tx.Where("status = ?", enummail.MailTypeSysMode).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select mail player list failed", err)
	}
	return out, nil
}

func CleanOverdueByCreateAt(tx *gorm.DB) error {
	if err := tx.Model(&mdmail.MailSendLog{}).
		Where("status < ? and created_at <  date_sub(curdate(),interval ? day)",
		enummail.MailSendOverdue,enummail.MailSendLogMaxNumber).
		UpdateColumn("status", enummail.MailSendOverdue).
		Error; err != nil {
		return errors.Internal("update player room play times failed", err)
	}
	if err := tx.Model(&mdmail.PlayerMail{}).
		Where("status < ? and created_at <  date_sub(curdate(),interval ? day)",
		enummail.PlayermailReadClose,enummail.PlayerMailMaxNumber).
		UpdateColumn("status", enummail.PlayermailOverdue).
		Error; err != nil {
		return errors.Internal("update player room play times failed", err)
	}
	return nil
}