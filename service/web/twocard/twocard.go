package twocard

import (
	pbtwo "playcards/proto/twocard"
	srvtwo "playcards/service/twocard/handler"
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

var rpc pbtwo.TwoCardSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllTwoCardMessage(brk); err != nil {
		return err
	}
	rpc = pbtwo.NewTwoCardSrvClient(
		apienum.TwoCardServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllTwoCardMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvtwo.TopicTwoCardGameStart),
		GameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtwo.TopicTwoCardSetBet),
		SetBetHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtwo.TopicTwoCardGameReady),
		GameReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtwo.TopicTwoCardGameSubmitCard),
		SubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtwo.TopicTwoCardGameResult),
		GameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomTwoCardExist),
		TwoCardExistHandle,
	)
	return nil
}

func GameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtwo.GameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgTwoCardGameStart, rs,enum.MsgTwoCardGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

func SetBetHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtwo.SetBetBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTwoCardSetBet, rs.Context,enum.MsgTwoCardSetBetCode)
	if err != nil {
		return err
	}
	return nil
}

func GameReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtwo.GameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgTwoCardGameReady, rs,enum.MsgTwoCardGameReadyCode)
	if err != nil {
		return err
	}
	return nil
}

func SubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtwo.SubmitCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTwoCardGameSubmitCard, rs.Context,enum.MsgTwoCardGameSubmitCardCode)
	if err != nil {
		return err
	}
	return nil
}

func GameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtwo.GameResultBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTwoCardGameResult, rs.Context,enum.MsgTwoCardGameResultCode)
	if err != nil {
		return err
	}
	return nil
}

func TwoCardExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	mdu := clients.GetClientByUserID(rs.UserID)
	if len(mdu) == 0{
		log.Err("twocard exist handle get user fail,uid:%d|%+v\n", rs.UserID, mdu)
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbtwo.RecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}

	reply, err := rpc.TwoCardRecovery(ctx, dr)
	if err != nil {
		log.Err("two card exist handle http err:%v|%v\n", rs, err)
		return err
	}
	topic := enum.MsgTwoCardExist
	topicCode := enum.MsgTwoCardExistCode
	if rs.RecoveryType == 1{
		topic = enum.MsgRoomSitDown
		topicCode = enum.MsgRoomSitDownCode
	}
	err = clients.SendTo(rs.UserID, t, topic, reply,topicCode)
	if err != nil {
		return err
	}
	return nil
}
