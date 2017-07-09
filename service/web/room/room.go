package room

import (
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/micro/protobuf/proto"
)

type RoomSrv struct {
	server server.Server
	broker broker.Broker
}

var RoomEvent = []string{
	srvroom.TopicRoomStatusChange,
	srvroom.TopicRoomReady,
	srvroom.TopicRoomUnReady,
	srvroom.TopicRoomJoin,
	srvroom.TopicRoomUnJoin,
}

func Init(brk broker.Broker) error {
	broker = brk
	if err := SubscribeRoomMessage(brk); err != nil {
		return err
	}
}

func SubscribeRoomMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomStatusChange),
		RoomStatusChangeHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomReady),
		RoomReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUnReady),
		RoomReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomJoin),
		RoomStatusChangeHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUnJoin),
		RoomStatusChangeHandler,
	)
}

func RoomStatusChangeHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomStatusChange, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	p := &pbroom.Position{}
	err := proto.Unmarshal(msg.Body, p)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomStatusChange, p)
	if err != nil {
		return err
	}
	return nil
}
