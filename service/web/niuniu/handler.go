package niuniu

import (
	pbniu "playcards/proto/niuniu"
	srvniu "playcards/service/niuniu/handler"
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
	if err := SubscribeAllNiuniuMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAllNiuniuMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuGameStart),
		NiuniuGameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuBeBanker),
		NiuniuBeBankerHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuSetBet),
		NiuniuSetBetHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuAllBet),
		NiuniuAllBetHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuGameReady),
		NiuniuGameReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuGameResult),
		NiuniuGameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
		TopicNiuniuCountDown),
		NiuniuCountDownHandler,
	)
	return nil
}

func NiuniuGameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.NiuniuGameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgNiuniuGameStart, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuBeBankerHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.BeBanker{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgNiuniuBeBanker, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuSetBetHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.SetBet{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgNiuniuSetBet, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuAllBetHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.AllBet{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgNiuniuAllBet, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuGameReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.GameReady{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgNiuniuGameReady, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuGameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.NiuniuRoomResult{}
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

func NiuniuCountDownHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbniu.CountDown{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgNiuniuCountDown, rs)
	if err != nil {
		return err
	}
	return nil
}
