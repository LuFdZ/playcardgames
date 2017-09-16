package handler

import (
	"playcards/model/activity"
	mda "playcards/model/activity/mod"
	pba "playcards/proto/activity"
	gctx "playcards/utils/context"
	utilproto "playcards/utils/proto"

	"golang.org/x/net/context"
)

type ActivitySrv struct {
}

func (*ActivitySrv) AddConfig(ctx context.Context, req *pba.ActivityConfig,
	rsp *pba.ActivityConfig) error {
	u := gctx.GetUser(ctx)
	cfg := mda.ActivityConfigFormProto(req)

	cfg.LastModifyUserID = u.UserID
	return activity.AddActivityConfig(cfg)
}

func (*ActivitySrv) DeleteConfig(ctx context.Context, req *pba.ActivityConfig,
	rsp *pba.ActivityConfig) error {
	return activity.DeleteActivityConfig(mda.ActivityConfigFormProto(req))
}

func (*ActivitySrv) ListConfig(ctx context.Context, req *pba.ActivityConfig,
	rsp *pba.ConfigList) error {
	aca, err := activity.ActivityConfigList()
	if err != nil {
		return err
	}

	return utilproto.ProtoSlice(aca, &rsp.List)
}

func (*ActivitySrv) UpdateConfig(ctx context.Context, req *pba.ActivityConfig,
	rsp *pba.ActivityConfig) error {
	u := gctx.GetUser(ctx)
	cfg := mda.ActivityConfigFormProto(req)

	cfg.LastModifyUserID = u.UserID
	return activity.UpdateActivityConfig(cfg)
}

func (*ActivitySrv) Invite(ctx context.Context, req *pba.InviteRequest,
	rsp *pba.InviteReply) error {
	u := gctx.GetUser(ctx)
	result, balances, err := activity.Invite(u, req.UserID)
	if err != nil {
		return err
	}
	var diamond int64
	if len(balances) > 1 {
		diamond = balances[0].Diamond
	}
	*rsp = pba.InviteReply{
		Result:  result,
		Diamond: diamond,
	}
	return nil
}

func (*ActivitySrv) Share(ctx context.Context, req *pba.ShareRequest,
	rsp *pba.ShareReply) error {
	u := gctx.GetUser(ctx)
	var diamond int64
	balance, err := activity.Share(u.UserID)
	if err != nil {
		// *rsp = pba.ShareReply{
		// 	Result:  2,
		// 	Diamond: diamond,
		// }
		return err
	}
	if balance != nil {
		diamond = balance.Diamond
	}
	*rsp = pba.ShareReply{
		Result:  1,
		Diamond: diamond,
	}
	return nil
}
