package room

import (
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
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
	if err := SubscribeAllRoomMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAllRoomMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomStatusChange),
		RoomStatusChangeHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomReady),
		RoomReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUnReady),
		RoomUnReadyHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomJoin),
		RoomJoin,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomUnJoin),
		RoomUnJoin,
	)
	return nil
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

func RoomJoin(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomRoomJoin, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomUnJoin(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomRoomUnJoin, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Position{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomReady, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomUnReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Position{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomUnReady, rs)
	if err != nil {
		return err
	}
	return nil
}

func AutoSubscribe(uid int32) {
	clients.AutoSubscribe(uid, RoomEvent)
}

func AutoUnSubscribe(uid int32) {
	clients.AutoUnSubscribe(uid, RoomEvent)
}
