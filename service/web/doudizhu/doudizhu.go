package doudizhu

import (
	apienum "playcards/service/api/enum"
	pbddz "playcards/proto/doudizhu"
	srvddz "playcards/service/doudizhu/handler"
	srvroom "playcards/service/room/handler"
	pbroom "playcards/proto/room"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/protobuf/proto"
	gctx "playcards/utils/context"
)

var rpc pbddz.DoudizhuSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllDoudizhuMessage(brk); err != nil {
		return err
	}
	rpc = pbddz.NewDoudizhuSrvClient(
		apienum.DoudizhuServiceName,
		client.DefaultClient,
	)

	return nil
}

func SubscribeAllDoudizhuMessage(brk broker.Broker) error {
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
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomDoudizhuExist),
		DoudizhuExistHandle,
	)
	return nil
}

func DoudizhuGameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.GameStart{}
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
	rs := &pbddz.BeBankerBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgDDZBeBanker, rs.Content)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuSubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.SubmitCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgDDZSubmitCard, rs.Content)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuGameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbddz.GameResultBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgDDZGameResult, rs.Content)
	if err != nil {
		return err
	}
	return nil
}

func DoudizhuExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbddz.DoudizhuRecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}
	reply, err := rpc.DoudizhuRecovery(ctx, dr)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgDoudizhuExist, reply)
	if err != nil {
		return err
	}
	return nil
}
