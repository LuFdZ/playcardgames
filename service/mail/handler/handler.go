package handler

import (
	"playcards/model/mail"
	mdgame "playcards/model/mail/mod"
	pbgame "playcards/proto/mail"
	enumgame "playcards/model/mail/enum"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"playcards/utils/auth"

	"golang.org/x/net/context"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
)

type MailSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *MailSrv {
	m := &MailSrv{
		server: s,
		broker: s.Options().Broker,
	}
	m.init()
	m.update(gt)
	return m
}

func (ms *MailSrv) init() {
	mail.RefreshAllMailInfoFromDB()
}

func (ms *MailSrv) update(gt *gsync.GlobalTimer) {
	//lock := "playcards.niu.update.lock"
}

func (ms *MailSrv) SendMail(ctx context.Context, req *pbgame.SendMailRequest,
	rsp *pbgame.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	msl, err := mail.SendMail(u.UserID,mdgame.SendMailLogFromProto(req.MailSend), req.Ids, req.SendAll)
	if err != nil {
		return err
	}
	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}

	msg := &pbgame.NewMailNoticeBro{
		Context: msl.ToProto(),
		Ids:     req.Ids,
	}
	topic.Publish(ms.broker, msg, TopicMailNotice)
	*rsp = *reply
	return nil
}

func (ms *MailSrv) RefreshAllMailInfoFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllMailInfoFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}

func (ms *MailSrv) RefreshAllMailSendLogFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllSendMailLogFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}

func (ms *MailSrv) RefreshAllPlayerMailFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllPlayerMailFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}
