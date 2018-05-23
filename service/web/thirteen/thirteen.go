package thirteen

import (
	pbthirteen "playcards/proto/thirteen"
	srvthirteen "playcards/service/thirteen/handler"
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

var rpc pbthirteen.ThirteenSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllThirteenMessage(brk); err != nil {
		return err
	}
	rpc = pbthirteen.NewThirteenSrvClient(
		apienum.ThirteenServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllThirteenMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
	TopicThirteenGameStart),
		ThirteenGameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
	TopicThirteenGameResult),
		ThirteenGameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
	TopicThirteenGameReady),
		ThirteenReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.
	TopicRoomThirteenExist),
		ThirteenExistHandle,
	)
	return nil
}

func ThirteenGameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GroupCard{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendTo(rs.UserID, t, enum.MsgThireteenGameStart, rs,enum.MsgThireteenGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

func ThirteenGameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GameResultList{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendToUsers(ids, t, enum.MsgThireteenGameResult, rs,enum.MsgThireteenGameResultCode)
	if err != nil {
		return err
	}
	return nil
}

func ThirteenReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GameReady{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendToUsers(ids, t, enum.MsgThireteenGameReady, rs,enum.MsgThireteenGameReadyCode)
	if err != nil {
		return err
	}
	return nil
}

func ThirteenExistHandle(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomExist{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	mdu := clients.GetClientByUserID(rs.UserID)
	if len(mdu) == 0{
		log.Err("thirteen exist handle get user fail,uid:%d|%+v\n", rs.UserID, mdu)
		return err
	}
	ctx := gctx.NewContext(clients.GetClientByUserID(rs.UserID)[0].Token())
	dr := &pbthirteen.ThirteenRecoveryRequest{
		UserID: rs.UserID,
		RoomID: rs.RoomID,
	}

	reply, err := rpc.ThirteenRecovery(ctx, dr)
	if err != nil {
		log.Err("thirteen exist handle http err:%v|%v\n", rs, err)
		return err
	}
	topic := enum.MsgThireteenExist
	topicCode := enum.MsgThireteenExistCode
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
