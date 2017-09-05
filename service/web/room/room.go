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
	// subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomDropped),
	// 	RoomDroppedHandler,
	// )
	//GiveUpGameResult
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

func RoomJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomUser{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	//fmt.Printf("RoomJoin:%d", rs.RoomID)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomJoin, rs)
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

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomUnJoin, rs)
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

	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomReady, rs)
	if err != nil {
		return err
	}
	return nil
}

func RoomUnReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomUser{}
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

func RoomResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomResults{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	//fmt.Printf("RoomResult:%+v", rs)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomResult, rs)
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
	//fmt.Printf("RoomGiveup:%+v", rs)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomGiveup, rs)
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
	//fmt.Printf("RoomGiveup:%+v", rs)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomGiveup, rs)
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
