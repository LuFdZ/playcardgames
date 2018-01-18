package club

import (
	cacheclub "playcards/model/club/cache"
	pbclub "playcards/proto/club"
	pbroom "playcards/proto/room"
	srvclub "playcards/service/club/handler"
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
	if err := SubscribeAllClubMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAllClubMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvclub.TopicClubMemberJoin),
		ClubMemberJoinHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvclub.TopicClubMemberLeave),
		ClubMemberLeaveHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvclub.TopicClubInfo),
		ClubInfoHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvclub.TopicClubOnlineStatus),
		ClubOnlineStatusHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicClubRoomCreate),
		ClubRoomCreateaHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicClubRoomFinish),
		ClubRoomFinishHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicClubRoomJoin),
		ClubRoomJoinHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicClubRoomUnJoin),
		ClubRoomUnJoinHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomGameStart),
		ClubRoomGameStartHandler,
	)

	return nil
}

func ClubMemberJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbclub.ClubMember{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}

	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubMemberJoin, rs, enum.MsgClubMemberJoinCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubMemberLeaveHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbclub.ClubMember{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubMemberLeave, rs, enum.MsgClubMemberLeaveCode)
	if err != nil {
		return err
	}

	return nil
}

func ClubOnlineStatusHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbclub.ClubMemberOnline{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = ClubOnlineStatus(t, rs)

	if err != nil {
		return nil
	}
	return nil
}

func ClubOnlineStatus(t string, rs *pbclub.ClubMemberOnline) error {
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubOnlineStatus, rs, enum.MsgClubOnlineStatusCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubInfoHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbclub.ClubInfo{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgClubInfo, rs, enum.MsgClubInfoCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubRoomCreateaHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubRoomCreate, rs, enum.MsgClubRoomCreateCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubRoomFinishHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubRoomFinish, rs, enum.MsgClubRoomFinishCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubRoomJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.ClubRoomUser{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubRoomJoin, rs, enum.MsgClubRoomJoinCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubRoomUnJoinHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubRoomUnJoin, rs, enum.MsgClubRoomUnJoinCode)
	if err != nil {
		return err
	}
	return nil
}

func ClubRoomGameStartHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.Room{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	err = clients.SendToUsers(uks, t, enum.MsgClubRoomGameStart, rs, enum.MsgClubRoomGameStartCode)
	if err != nil {
		return err
	}
	return nil
}

