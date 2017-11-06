package notice

import (
	dbn "playcards/model/notice/db"
	mdn "playcards/model/notice/mod"
	mdpage "playcards/model/page"
	"playcards/utils/db"
	"playcards/utils/errors"
)

func CreateNotice(n *mdn.Notice) (*mdn.Notice, error) {
	return dbn.CreateNotice(db.DB(), n)
}
func UpdateNotice(n *mdn.Notice) (*mdn.Notice, error) {
	if n.NoticeID <1{
		return nil,errors.Conflict(80001, "未找到数据ID！")
	}
	return dbn.UpdateNotice(db.DB(), n)
}

func GetNotice(channel string, versions string) (*mdn.NoticeList, error) {
	return dbn.GetNotice(db.DB(), channel, versions)
}

func AllNotice() (*mdn.NoticeList, error) {
	return dbn.AllNotice(db.DB())
}

func PageNoticeList(page *mdpage.PageOption, n *mdn.Notice) (
	[]*mdn.Notice, int64, error) {
	return dbn.PageNoticeList(db.DB(), page, n)
}
