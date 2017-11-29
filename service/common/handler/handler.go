package handler

import (
	"playcards/model/common"
	mdcommon "playcards/model/common/mod"
	mdpage "playcards/model/page"
	pbcommon "playcards/proto/common"
	gctx "playcards/utils/context"
	utilproto "playcards/utils/proto"

	"golang.org/x/net/context"
)

type CommonSrv struct {
}

func NewHandler() *CommonSrv {
	cs := &CommonSrv{}
	cs.Init()
	return cs
}

func (cs *CommonSrv) Init() {
	common.RefreshAllFromDB()
}

func (cs *CommonSrv) CreateBlackList(ctx context.Context,
	req *pbcommon.BlackList, rsp *pbcommon.CommonReply) error {
	u := gctx.GetUser(ctx)
	err := common.CreateBlackList(req.Type, req.OriginID, req.TargetID, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbcommon.CommonReply{
		Result: 1,
	}
	return nil
}

func (cs *CommonSrv) CreateExamine(ctx context.Context,
	req *pbcommon.Examine, rsp *pbcommon.CommonReply) error {
	u := gctx.GetUser(ctx)
	me := mdcommon.ExamineFromProto(req)
	err := common.CreateExamine(me, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbcommon.CommonReply{
		Result: 1,
	}
	return nil
}

func (cs *CommonSrv) CancelBlackList(ctx context.Context,
	req *pbcommon.BlackList, rsp *pbcommon.CommonReply) error {
	u := gctx.GetUser(ctx)
	mBl := mdcommon.BlackListFromProto(req)
	err := common.CancelBlackList(mBl, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbcommon.CommonReply{
		Result: 1,
	}
	return nil
}

func (cs *CommonSrv) UpdateExamine(ctx context.Context,
	req *pbcommon.Examine, rsp *pbcommon.CommonReply) error {
	u := gctx.GetUser(ctx)
	//me := mdcommon.ExamineFromProto(req)
	err := common.UpdateExamine(req.Type, req.AuditorID, req.ApplicantID, req.Status, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbcommon.CommonReply{
		Result: 1,
	}
	return nil
}

func (cs *CommonSrv) PageBlackList(ctx context.Context,
	req *pbcommon.PageBlackListRequest, rsp *pbcommon.PageBlackListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := common.PageBlackList(page,
		mdcommon.BlackListFromProto(req.BlackList))
	if err != nil {
		return err
	}

	err = utilproto.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *CommonSrv) PageExamine(ctx context.Context,
	req *pbcommon.PageExamineRequest, rsp *pbcommon.PageExamineReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := common.PageExamine(page,
		mdcommon.ExamineFromProto(req.Examine))
	if err != nil {
		return err
	}
	err = utilproto.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *CommonSrv) RefreshAllFromDB(ctx context.Context,
	req *pbcommon.BlackList, rsp *pbcommon.CommonReply) error {
	err := common.RefreshAllFromDB()
	if err != nil {
		return err
	}
	*rsp = pbcommon.CommonReply{
		Result: 1,
	}
	return nil
}
