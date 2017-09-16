package db

import (
	mdpage "playcards/model/page"
	//errr "playcards/model/notic/errors"

	enum "playcards/model/notice/enum"
	mdn "playcards/model/notice/mod"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateNotice(tx *gorm.DB, n *mdn.Notice) (*mdn.Notice, error) {
	if err := tx.Create(n).Error; err != nil {
		return nil, errors.Internal("create notice failed", err)
	}
	return n, nil
}

func UpdateNotice(tx *gorm.DB, n *mdn.Notice) (*mdn.Notice, error) {
	now := gorm.NowFunc()
	// notice := &mdn.Notice{
	// 	NoticeContent: n.NoticeContent,
	// 	Description:   n.Description,
	// 	Param:         n.Param,
	// 	StartAt:       n.StartAt,
	// 	EndAt:         n.EndAt,
	// 	UpdatedAt:     &now,
	// }
	n.UpdatedAt = &now
	if err := tx.Model(n).Updates(n).Error; err != nil {
		return nil, errors.Internal("update notice failed", err)
	}
	return n, nil
}

func GetNotice(tx *gorm.DB, channel string, versions string) (*mdn.NoticeList, error) {
	var (
		notices []*mdn.Notice
		t       = gorm.NowFunc()
	)
	sqlstr := " start_at < ? and end_at > ? and status = ? and" +
		" (case when CHAR_LENGTH(channels)>1 then FIND_IN_SET(?,channels) else true end)" +
		"and (case when CHAR_LENGTH(versions)>1 then FIND_IN_SET(?,versions) else true end)"
	if err := tx.Where(sqlstr, t, t, enum.NoticeStatusable,
		channel, versions,
	).Find(&notices).Error; err != nil {
		return nil, errors.Internal("get notice failed", err)
	}
	out := &mdn.NoticeList{
		List: notices,
	}
	return out, nil
}

func AllNotice(tx *gorm.DB) (*mdn.NoticeList, error) {
	var (
		notices []*mdn.Notice
	)
	if err := tx.Order("created_at").Find(&notices).Error; err != nil {
		return nil, errors.Internal("notice list failed", err)
	}

	out := &mdn.NoticeList{
		List: notices,
	}
	return out, nil
}

func PageNoticeList(tx *gorm.DB, page *mdpage.PageOption,
	n *mdn.Notice) ([]*mdn.Notice, int64, error) {
	var out []*mdn.Notice
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page notice failed", rtx.Error)
	}
	return out, rows, nil
}
