package fourcard

import (
	pbtow "playcards/proto/towcard"
	srvtow "playcards/service/towcard/handler"
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

var rpc pbtow.TowCardSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllFourCardMessage(brk); err != nil {
		return err
	}
	rpc = pbtow.NewTowCardSrvClient(
		apienum.FourCardServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllFourCardMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvtow.TopicTowCardGameStart),
		GameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtow.TopicTowCardSetBet),
		SetBetHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtow.TopicTowCardGameReady),
		GameReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtow.TopicTowCardGameSubmitCard),
		SubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvtow.TopicTowCardGameResult),
		GameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomTowCardExist),
		TowCardExistHandle,
	)
	return nil
}

func GameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtow.GameStart{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgTowCardGameStart, rs,enum.MsgTowCardGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

func SetBetHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtow.SetBetBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTowCardSetBet, rs.Context,enum.MsgTowCardSetBetCode)
	if err != nil {
		return err
	}
	return nil
}

func GameReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtow.GameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgTowCardGameReady, rs,enum.MsgTowCardGameReadyCode)
	if err != nil {
		return err
	}
	return nil
}

func SubmitCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtow.SubmitCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTowCardGameSubmitCard, rs.Context,enum.MsgTowCardGameSubmitCardCode)
	if err != nil {
		return err
	}
	return nil
}

func GameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbtow.GameResultBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(rs.Ids, t, enum.MsgTowCardGameResult, rs.Context,enum.MsgTowCardGameResultCode)
	if err != nil {
		return err
	}
	return nil
}

func TowCardExistHandle(p broker.Publication) error {

	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbtow.RecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}
	reply, err := rpc.TowCardRecovery(ctx, dr)
	if err != nil {
		log.Err("tow card exist handle http err:%v|%v\n", rs, err)
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgTowCardExist, reply,enum.MsgTowCardExistCode)
	if err != nil {
		return err
	}
	return nil
}
