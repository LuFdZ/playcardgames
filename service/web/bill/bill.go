package bill

import (
	pbbill "playcards/proto/bill"
	srvbill "playcards/service/bill/handler"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/protobuf/proto"
)

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllBillMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAllBillMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvbill.
	TopicBillChange),
		BillChangeHandler,
	)
	return nil
}

func BillChangeHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbbill.BalanceChange{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgBillChange, rs)
	if err != nil {
		return err
	}
	return nil
}
