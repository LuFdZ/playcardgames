package handler

import (
	"playcards/model/activity"
	pba "playcards/proto/activity"
	gctx "playcards/utils/context"
	pbmail "playcards/proto/mail"
	srvmail "playcards/service/mail/handler"
	enummail "playcards/model/mail/enum"
	"playcards/utils/topic"
	"playcards/utils/auth"
	"golang.org/x/net/context"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"playcards/utils/tools"
)

type ActivitySrv struct {
	server server.Server
	broker broker.Broker
}

func NewHandler(s server.Server) *ActivitySrv {
	b := &ActivitySrv{
		server: s,
		broker: s.Options().Broker,
	}
	return b
}

func (as *ActivitySrv) Invite(ctx context.Context, req *pba.InviteRequest,
	rsp *pba.InviteReply) error {
	u := gctx.GetUser(ctx)
	result, err := activity.Invite(u, req.UserID)
	//fmt.Printf("Invite Err:%+v", err)
	if err != nil {
		return err
	}
	//if result == 2 {
	//	return errors.ErrInviterOverdue
	//}
	//var diamond int64
	//if len(balances) > 1 {
	//	diamond = balances[0].Balance
	//}

	if result == 1{
		mailReqA := &pbmail.SendSysMailRequest{
			MailID: enummail.MailBeInvite,
			Ids:    []int32{req.UserID},
			Args:   []string{tools.IntToString(u.UserID)},
		}
		topic.Publish(as.broker, mailReqA, srvmail.TopicSendSysMail)

		mailReqB := &pbmail.SendSysMailRequest{
			MailID: enummail.MailInvite,
			Ids:    []int32{u.UserID},
			Args:   []string{tools.IntToString(req.UserID)},
		}
		topic.Publish(as.broker, mailReqB, srvmail.TopicSendSysMail)
	}else if result == 2{
		mailReq := &pbmail.SendSysMailRequest{
			MailID: enummail.MailInviteOverdue,
			Ids:    []int32{u.UserID},
			Args:   nil,
		}
		topic.Publish(as.broker, mailReq, srvmail.TopicSendSysMail)
	}
	*rsp = pba.InviteReply{
		Result:  result,
		//Diamond: diamond,
		InviteID:req.UserID,
	}
	return nil
}

func (as *ActivitySrv) Share(ctx context.Context, req *pba.ShareRequest,
	rsp *pba.ShareReply) error {
	u := gctx.GetUser(ctx)
	//var diamond int64
	err := activity.Share(u.UserID) //balance,
	if err != nil {
		// *rsp = pba.ShareReply{
		// 	Result:  2,
		// 	Diamond: diamond,
		// }
		return err
	}
	//if balance != nil {
	//	diamond = balance.Balance
	//}
	*rsp = pba.ShareReply{
		Result:  1,
		Diamond: 0,//diamond,
	}
	mailReq := &pbmail.SendSysMailRequest{
		MailID: enummail.MailShare,
		Ids:    []int32{u.UserID},
		Args:   nil,
	}
	topic.Publish(as.broker, mailReq, srvmail.TopicSendSysMail)
	//fmt.Printf("AAAAAAAShare:%v",rsp)
	return nil
}

func (as *ActivitySrv) InviteUserInfo(ctx context.Context, req *pba.InviteRequest,
	rsp *pba.InviteUserInfoReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	count, err := activity.InviteUserInfo(u.UserID)
	if err != nil{
		return nil
	}
	*rsp = pba.InviteUserInfoReply{
		InviteUserID:u.InviteUserID,
		Count: count,
	}
	return nil
}
