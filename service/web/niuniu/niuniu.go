package niuniu

import (
	pbniu "playcards/proto/niuniu"
	srvniu "playcards/service/niuniu/handler"
	srvroom "playcards/service/room/handler"
	pbroom "playcards/proto/room"
	apienum "playcards/service/api/enum"
	"github.com/micro/go-micro/client"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/protobuf/proto"
	gctx "playcards/utils/context"
)

var rpc pbniu.NiuniuSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllNiuniuMessage(brk); err != nil {
		return err
	}
	rpc = pbniu.NewNiuniuSrvClient(
		apienum.NiuniuServiceName,
		client.DefaultClient,
	)
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
	//subscribe.SrvSubscribe(brk, topic.Topic(srvniu.
	//TopicNiuniuCountDown),
	//	NiuniuCountDownHandler,
	//)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomNiuniuExist),
		NiuniuExistHandle,
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
	err = clients.SendToUsers(ids, t, enum.MsgNiuniuBeBanker, rs)
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
	err = clients.SendToUsers(ids, t, enum.MsgNiuniuSetBet, rs)
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
	err = clients.SendToUsers(ids, t, enum.MsgNiuniuGameReady, rs)
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
	err = clients.SendToUsers(ids, t, enum.MsgNiuniuGameResult, rs)
	if err != nil {
		return err
	}
	return nil
}

func NiuniuExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbniu.NiuniuRecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}
	reply, err := rpc.NiuniuRecovery(ctx, dr)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgNiuniuExist, reply)
	if err != nil {
		return err
	}
	return nil
}