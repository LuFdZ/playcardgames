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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubMemberJoin, rs)
		if err != nil {
			return err
		}
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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubMemberLeave, rs)
		if err != nil {
			return err
		}
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

func ClubOnlineStatus(topic string, rs *pbclub.ClubMemberOnline) error {
	uks, err := cacheclub.ListClubMemberHKey(rs.ClubID, true)
	if err != nil {
		return err
	}
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, topic, enum.MsgClubOnlineStatus, rs)
		if err != nil {
			return err
		}
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
	err = clients.SendTo(rs.UserID, t, enum.MsgClubInfo, rs)
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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubRoomCreate, rs)
		if err != nil {
			return err
		}
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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubRoomFinish, rs)
		if err != nil {
			return err
		}
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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubRoomJoin, rs)
		if err != nil {
			return err
		}
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
	for _, uid := range uks {
		err = clients.SendToNoLog(uid, t, enum.MsgClubRoomUnJoin, rs)
		if err != nil {
			return err
		}
	}
	return nil
}
