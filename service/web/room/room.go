package room

import (
	apienum "gdc/service/api/enum"
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/protobuf/proto"
)

var rpc pbroom.RoomSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllRoomMessage(brk); err != nil {
		return err
	}

	rpc = pbroom.NewRoomSrvClient(
		apienum.RoomServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllRoomMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomCreate),
		RoomCreateHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomReady),
		RoomReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomJoin),
		RoomJoinHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUnJoin),
		RoomUnJoinHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomResult),
		RoomResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomGiveup),
		RoomGiveupHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomShock),
		RoomShockHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUserConnection),
		UserConnectionHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomRenewal),
		RoomRenewalHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomVoiceChat),
		RoomVoiceChatHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomExist),
		RoomExistHandler,
	)
	return nil
}

func RoomCreateHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	//fmt.Printf("AAAAAAAAA:%d",rs.UserID)
	err = clients.SendTo(rs.UserID, t, enum.MsgRoomCreate, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomUser{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	//fmt.Printf("RoomJoinHandler:%v",ids)
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomJoin, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomUnJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomUser{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	//fmt.Printf("RoomUnJoinHandler:%v",ids)
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomUnJoin, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomUser{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	//fmt.Printf("RoomReadyHandler:%v\n",ids)
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomReady, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomResults{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomResult, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomGiveupHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.GiveUpGameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomGiveup, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomShockHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Shock{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserIDTo, t, enum.MsgRoomShock, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomRenewalHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RenewalRoomReady{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomRenewal, rs)
	if err != nil {
		return err
	}
	return nil
}

func UserConnectionHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.UserConnection{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomUserConnection, rs)
	if err != nil {
		return err
	}

	return nil
}

func RoomVoiceChatHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.VoiceChat{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendRoomUsers(ids, t, enum.MsgRoomVoiceChat, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomExistHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.CheckRoomExistReply{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgRoomExist, rs)
	if err != nil {
		return err
	}
	return nil
}

