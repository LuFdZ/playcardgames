package handler

import (
	"playcards/model/activity"
	"playcards/model/activity/errors"
	pba "playcards/proto/activity"
	"playcards/utils/auth"
	gctx "playcards/utils/context"

	"golang.org/x/net/context"
)

type ActivitySrv struct {
}

func (*ActivitySrv) Invite(ctx context.Context, req *pba.InviteRequest,
	rsp *pba.InviteReply) error {
	u := gctx.GetUser(ctx)
	result, balances, err := activity.Invite(u, req.UserID)
	//fmt.Printf("Invite Err:%+v", err)
	if err != nil {
		return err
	}
	if result == 2 {
		return errors.ErrInviterOverdue
	}
	var diamond int64
	if len(balances) > 1 {
		diamond = balances[0].Balance
	}
	*rsp = pba.InviteReply{
		Result:  result,
		Diamond: diamond,
		InviteID:req.UserID,
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
		diamond = balance.Balance
	}
	*rsp = pba.ShareReply{
		Result:  1,
		Diamond: diamond,
	}
	//fmt.Printf("AAAAAAAShare:%v",rsp)
	return nil
}

func (*ActivitySrv) InviteUserInfo(ctx context.Context, req *pba.InviteRequest,
	rsp *pba.InviteUserInfoReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	count, err := activity.InviteUserInfo(u.UserID)
	*rsp = pba.InviteUserInfoReply{
		InviteUserID:u.InviteUserID,
		Count: count,
	}
	return nil
}
