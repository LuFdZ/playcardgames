package handler

import (
	"playcards/model/bill"
	_ "playcards/model/bill"
	enumbill "playcards/model/bill/enum"
	pbbill "playcards/proto/bill"
	"playcards/utils/auth"
	gctx "playcards/utils/context"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"


	"golang.org/x/net/context"
)

type BillSrv struct {
	server server.Server
	broker broker.Broker
}

func NewHandler(s server.Server) *BillSrv {
	b := &BillSrv{
		server: s,
		broker: s.Options().Broker,
	}
	return b
}

func (b *BillSrv) GetBalance(ctx context.Context,
	req *pbbill.GetBalanceRequest, rsp *pbbill.Balance) error {
	u := gctx.GetUser(ctx)
	uid := req.UserID
	err := auth.Check(u.Rights, auth.RightsBillView)
	if err != nil || uid <= 0 {
		uid = u.UserID
	}
	balance, err := bill.GetAllUserBalance(uid)
	if err != nil {
		return err
	}

	*rsp = *balance
	return nil
}

func (b *BillSrv) GetUserBalance(ctx context.Context,
	req *pbbill.GetBalanceRequest, rsp *pbbill.Balance) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	balance, err := bill.GetAllUserBalance(u.UserID)
	if err != nil {
		return err
	}
	*rsp = *balance
	return nil
}

//func (b *BillSrv) GainBalance(ctx context.Context, req *pbbill.Balance,
//	rsp *pbbill.Balance) error {
//	u := gctx.GetUser(ctx)
//	ub, err := bill.GainBalance(req.UserID, u.UserID,
//		&mbill.Balance{Gold: req.Gold, Diamond: req.Diamond})
//	if err != nil {
//		return err
//	}
//	*rsp = *ub.ToProto()
//	// if req.UserID != u.UserID {
//	// 	msg := &pbbill.BalanceChange{
//	// 		UserID:  req.UserID,
//	// 		Diamond: ub.Diamond,
//	// 	}
//	// 	topic.Publish(b.broker, msg, TopicBillChange)
//	// }
//	return nil
//}

func (b *BillSrv) Recharge(ctx context.Context, req *pbbill.RechargeRequest,
	rsp *pbbill.RechargeReply) error {
	rsp.Result = 2
	rsp.Code = 101
	u := gctx.GetUser(ctx)
	//orderID, err := strconv.ParseInt(req.OrderID, 10, 64)
	//if err != nil {
	//	rsp.Code = 102
	//	return err
	//}
	res, ub, err := bill.Recharge(req.UserID, u.UserID, req.Diamond,
		req.OrderID, enumbill.JournalTypeRecharge, u.Channel)
	if err != nil {
		rsp.Code = 102
		return err
	}
	if res == 1 {
		rsp.Code = 0
		rsp.Result = 1

		msg := &pbbill.BalanceChange{
			UserID:  req.UserID,
			Diamond: ub.Balance,
		}
		topic.Publish(b.broker, msg, TopicBillChange)
	}
	//rsp.Result = result
	return nil
}
