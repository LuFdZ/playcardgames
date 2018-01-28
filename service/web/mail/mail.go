package mail

import (
	pbmail "playcards/proto/mail"
	srvmail "playcards/service/mail/handler"
	apienum "playcards/service/api/enum"
	"github.com/micro/go-micro/client"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/protobuf/proto"
	"playcards/utils/log"
)

var rpc pbmail.MailSrvClient

func Init(brk broker.Broker) error {
	if err := SubscribeAllMailMessage(brk); err != nil {
		return err
	}
	rpc = pbmail.NewMailSrvClient(
		apienum.NiuniuServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllMailMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvmail.TopicSendMail),
		NewMailNoticeHandler,
	)
	return nil
}

func NewMailNoticeHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbmail.NewMailNoticeBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	if rs.Ids != nil {
		err = clients.SendToUsers(rs.Ids, t, enum.MsgNewMailNotice, rs.Context.SendLogID, enum.MsgNewMailNoticeCode)
		if err != nil {
			return err
		}
	}else{
		err = clients.Send(t, enum.MsgNewMailNotice, rs.Context.SendLogID, enum.MsgNewMailNoticeCode)
		if err != nil {
			return err
		}
	}
	if err != nil {
		log.Err("new mail notice handler http err:%v|%v\n", rs, err)
		return err
	}
	return nil
}
