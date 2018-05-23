package twocard

import (
	pbrun "playcards/proto/runcard"
	srvrun "playcards/service/runcard/handler"
	srvroom "playcards/service/room/handler"
	pbroom "playcards/proto/room"
	apienum "playcards/service/api/enum"
	"github.com/micro/go-micro/client"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"
	"playcards/utils/log"
	"github.com/micro/go-micro/broker"
	"github.com/micro/protobuf/proto"
	gctx "playcards/utils/context"
)

var rpc pbrun.RunCardSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllRunCardMessage(brk); err != nil {
		return err
	}
	rpc = pbrun.NewRunCardSrvClient(
		apienum.RunCardServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllRunCardMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvrun.TopicRunCardGameStart),
		GameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvrun.TopicRunCardGameSubmitCard),
		SubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvrun.TopicRunCardGameResult),
		GameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomRunCardExist),
		RunCardExistHandle,
	)
	return nil
}

func GameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbrun.GameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgRunCardGameStart, rs, enum.MsgRunCardGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

func SubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbrun.SubmitCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgRunCardSubmitCard, rs.Context, enum.MsgRunCardSubmitCardCode)
	if err != nil {
		return err
	}
	return nil
}

func GameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbrun.GameResultBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgRunCardGameResult, rs.Context, enum.MsgRunCardGameResultCode)
	if err != nil {
		return err
	}
	return nil
}

func RunCardExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	mdu := clients.GetClientByUserID(rs.UserID)
	if len(mdu) == 0 {
		log.Err("run card exist handle get user fail,uid:%d|%+v\n", rs.UserID, mdu)
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbrun.RecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}

	reply, err := rpc.RunCardRecovery(ctx, dr)
	if err != nil {
		log.Err("run card exist handle http err:%v|%v\n", rs, err)
		return err
	}
	topic := enum.MsgRunCardExist
	topicCode := enum.MsgRunCardExistCode
	if rs.RecoveryType == 1 {
		topic = enum.MsgRoomSitDown
		topicCode = enum.MsgRoomSitDownCode
	}
	err = clients.SendTo(rs.UserID, t, topic, reply, topicCode)
	if err != nil {
		return err
	}
	return nil
}
