package fourcard

import (
	pbfour "playcards/proto/fourcard"
	srvfour "playcards/service/fourcard/handler"
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

var rpc pbfour.FourCardSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllFourCardMessage(brk); err != nil {
		return err
	}
	rpc = pbfour.NewFourCardSrvClient(
		apienum.FourCardServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllFourCardMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvfour.TopicFourCardGameStart),
		GameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvfour.TopicFourCardSetBet),
		SetBetHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvfour.TopicFourCardGameReady),
		GameReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvfour.TopicFourCardGameSubmitCard),
		SubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvfour.TopicFourCardGameResult),
		GameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomFourCardExist),
		FourCardExistHandle,
	)
	return nil
}

func GameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbfour.GameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgFourCardGameStart, rs,enum.MsgFourCardGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

func SetBetHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbfour.SetBetBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgFourCardSetBet, rs.Context,enum.MsgFourCardSetBetCode)
	if err != nil {
		return err
	}
	return nil
}

func GameReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbfour.GameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgFourCardGameReady, rs,enum.MsgFourCardGameReadyCode)
	if err != nil {
		return err
	}
	return nil
}

func SubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbfour.SubmitCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgFourCardGameSubmitCard, rs.Context,enum.MsgFourCardGameSubmitCardCode)
	if err != nil {
		return err
	}
	return nil
}

func GameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbfour.GameResultBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgFourCardGameResult, rs.Context,enum.MsgFourCardGameResultCode)
	if err != nil {
		return err
	}
	return nil
}

func FourCardExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbfour.RecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}
	reply, err := rpc.FourCardRecovery(ctx, dr)
	if err != nil {
		log.Err("four card exist handle http err:%v|%v\n", rs, err)
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgFourCardExist, reply,enum.MsgFourCardExistCode)
	if err != nil {
		return err
	}
	return nil
}
