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
