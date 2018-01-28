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
	return msl, nil
}

func GetAllUser(tx *gorm.DB) []int32 {
	var (
		ids []int32
	)
	sql, param, err := squirrel.
	Select(" user_id").From("users").ToSql()
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
	now := gorm.NowFunc()
	var count int32 = 0
	for _, mdpm := range mdpms {
		mdpm.UpdatedAt = &now
		mdpm.CreatedAt = &now
		if err := tx.Create(mdpm).Error; err != nil {
			log.Err("create player mail failed! uid:%d,send_log_id:%d,err:%+v", mdpm.UserID, mdpm.SendLogID, err)
		}
		if err := cachemail.SetPlayerMail(mdpm).Error; err != nil {
			log.Err("create player mail redis failed! uid:%d,send_log_id:%d,err:%+v", mdpm.UserID, mdpm.SendLogID, err)
		}
		count ++
	}
	mdmsl.SendCount = count
	mdmsl.TotalCount = int32(len(mdpms))
	mdmsl, _ = UpdateMailSendLog(tx, mdmsl)
	if err := cachemail.SetMailSendLog(mdmsl).Error; err != nil {
		log.Err("create mail send log redis failed! send log ID:%d,err:%+v", mdmsl.SendLogID, err)
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
