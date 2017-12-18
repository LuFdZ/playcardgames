package niuniu

import (
	pbddz "playcards/proto/doudizhu"
	srvddz "playcards/service/doudizhu/handler"
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
	if err := SubscribeAlldoudizhuMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAlldoudizhuMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZGameStart),
		DoudizhuGameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZBeBanker),
		DoudizhuBeBankerHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZSubmitCard),
		DoudizhuSubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZGameResult),
		DoudizhuGameResultHandler,
	)
	return nil
}

func DoudizhuGameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.DDZGameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgDDZGameStart, rs)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuBeBankerHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.BeBanker{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgDDZBeBanker, rs)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuSubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.SubmitCard{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID,t, enum.MsgDDZSubmitCard, rs)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuGameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.DDZGameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgNiuniuGameResult, rs)
	if err != nil {
		return err
	}
	return nil
}
