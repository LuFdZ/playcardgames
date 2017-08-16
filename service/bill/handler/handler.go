package handler

import (
	"playcards/model/bill"
	_ "playcards/model/bill"
	enumbill "playcards/model/bill/enum"
	mbill "playcards/model/bill/mod"
	pbbill "playcards/proto/bill"
	"playcards/utils/auth"
	gctx "playcards/utils/context"

	"golang.org/x/net/context"
)

type BillSrv struct {
}

func (b *BillSrv) GetBalance(ctx context.Context,
	req *pbbill.GetBalanceRequest, rsp *pbbill.Balance) error {
	u := gctx.GetUser(ctx)
	uid := req.UserID
	err := auth.Check(u.Rights, auth.RightsBillView)
	if err != nil || uid <= 0 {
		uid = u.UserID
	}
	balance, err := bill.GetUserBalance(uid)
	if err != nil {
		return err
	}
	*rsp = *balance.ToProto()
	return nil
}

func (b *BillSrv) GainBalance(ctx context.Context, req *pbbill.Balance,
	rsp *pbbill.Balance) error {
	u := gctx.GetUser(ctx)
	ub, err := bill.GainBalance(req.UserID, u.UserID,
		&mbill.Balance{Gold: req.Gold, Diamond: req.Diamond})
	if err != nil {
		return err
	}
	*rsp = *ub.ToProto()
	return nil
}

func (b *BillSrv) Recharge(ctx context.Context, req *pbbill.RechargeRequest,
	rsp *pbbill.RechargeReply) error {
	rsp.Result = 2
	rsp.Code = 101
	u := gctx.GetUser(ctx)
	res, err := bill.Recharge(req.UserID, u.UserID, (int64)(req.Diamond),
		req.OrderID, enumbill.JournalTypeRecharge)
	if err != nil {
		rsp.Code = 102
		return err
	}
	if res == 1 {
		rsp.Code = 0
		rsp.Result = 1
	}

	//rsp.Result = result
	return nil
}
