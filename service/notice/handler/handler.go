package handler

import (
	"playcards/model/notice"
	mdn "playcards/model/notice/mod"
	mdpage "playcards/model/page"
	pbn "playcards/proto/notice"
	"playcards/utils/auth"
	utilpb "playcards/utils/proto"

	"golang.org/x/net/context"
)

type NoticeSrv struct {
}

func (us *NoticeSrv) GetNotice(ctx context.Context, req *pbn.Notice,
	rsp *pbn.NoticeListReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	nl, err := notice.GetNotice(req.Versions, u.Channel)
	if err != nil {
		return err
	}
	*rsp = *nl.ToProto()

	return nil
}

func (us *NoticeSrv) AllNotice(ctx context.Context, req *pbn.Notice,
	rsp *pbn.NoticeListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	nl, err := notice.AllNotice()
	if err != nil {
		return err
	}
	*rsp = *nl.ToProto()

	return nil
}

func (us *NoticeSrv) CreateNotice(ctx context.Context, req *pbn.Notice,
	rsp *pbn.Notice) error {

	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	n, err := notice.CreateNotice(mdn.NoticeFromProto(req))
	if err != nil {
		return err
	}
	*rsp = *n.ToProto()

	return nil
}

func (us *NoticeSrv) UpdateNotice(ctx context.Context, req *pbn.Notice,
	rsp *pbn.Notice) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	n, err := notice.UpdateNotice(mdn.NoticeFromProto(req))
	if err != nil {
		return err
	}
	*rsp = *n.ToProto()

	return nil
}

func (rs *NoticeSrv) PageNoticeList(ctx context.Context,
	req *pbn.PageNoticeListRequest, rsp *pbn.PageNoticeListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	l, rows, err := notice.PageNoticeList(page,
		mdn.NoticeFromProto(req.Notice))
	if err != nil {
		return err
	}

	err = utilpb.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	return nil
}
