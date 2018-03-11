package room

import (
	apienum "playcards/service/api/enum"
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	cacheroom "playcards/model/room/cache"
	"playcards/service/web/clients"
	"playcards/service/web/enum"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/protobuf/proto"
	"playcards/utils/log"
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
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomNotice),
		RoomNoticeHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomShuffleCardBro),
		ShuffleCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicRoomChat),
		RoomChatHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvroom.TopicBankerList),
		BankerListHandler,
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
	err = clients.SendTo(rs.UserID, t, enum.MsgRoomCreate, rs,enum.MsgRoomCreateCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomJoin, rs,enum.MsgRoomJoinCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomUnJoin, rs,enum.MsgRoomUnJoinCode)
	if err != nil {
		return err
	}
	err = clients.SendTo(rs.UserID, t, enum.MsgRoomUnJoin, rs,enum.MsgRoomUnJoinCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomReady, rs,enum.MsgRoomReadyCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomResult, rs,enum.MsgRoomResultCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomGiveup, rs,enum.MsgRoomGiveupCode,)
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
	err = clients.SendTo(rs.UserIDTo, t, enum.MsgRoomShock, rs,enum.MsgRoomShockCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomRenewal, rs,enum.MsgRoomRenewalCode)
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
	err = clients.SendToUsers(ids, t, enum.MsgRoomVoiceChat, rs,enum.MsgRoomVoiceChatCode)
	if err != nil {
		return err
	}
	return nil
}

func RoomChatHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomChat{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendToUsers(ids, t, enum.MsgRoomChat, rs,enum.MsgRoomChatCode)
	if err != nil {
		return err
	}
	return nil
}

//func RoomExistHandler(p broker.Publication) error {
//	t := p.Topic()
//	msg := p.Message()
//	rs := &pbroom.CheckRoomExistReply{}
//	err := proto.Unmarshal(msg.Body, rs)
//	if err != nil {
//		return err
//	}
//	err = clients.SendTo(rs.UserID, t, enum.MsgRoomExist, rs)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func RoomNoticeHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.RoomNotice{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendToUsers(ids, t, enum.MsgRoomNotice, rs,enum.MsgRoomNoticeCode)
	if err != nil {
		return err
	}
	return nil
}

//func UserConnectionHandler(p broker.Publication) error {
//	t := p.Topic()
//	msg := p.Message()
//	rs := &pbroom.UserConnection{}
//	err := proto.Unmarshal(msg.Body, rs)
//	if err != nil {
//		return err
//	}
//	ids := rs.Ids
//	rs.Ids = nil
//	err = clients.SendRoomUsers(ids, t, enum.MsgRoomUserConnection, rs)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func UserConnectionHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.UserConnection{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	inRoom := cacheroom.ExistRoomUser(rs.UserID)
	if inRoom {
		mdroom, err := cacheroom.GetRoomUserID(rs.UserID)
		if err != nil {
			log.Err("UserConnectionHandlerErr uid:%d,status:%d,err:%d", rs.UserID, rs.Status, err)
			return err
		}
		rs := &pbroom.UserConnection{rs.UserID, rs.Status, mdroom.Ids}
		var ids []int32
		for _,id := range rs.Ids{
			if id != rs.UserID{
				ids = append(ids,id)
			}
		}
		if len(ids) == 0{
			return nil
		}
		rs.Ids = nil
		err = clients.SendToUsers(ids, t, enum.MsgRoomUserConnection, rs,enum.MsgRoomUserConnectionCode)
		if err != nil {
			return err
		}
		//if rs.Status == enum.SocketAline {
		//	rs := &pbroom.RoomExist{rs.UserID, mdroom.RoomID, mdroom.GameType}
		//	topic.Publish(brok, rs, srvroom.TopicRoomExist)
		//}
	}
	return nil
}

func ShuffleCardHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.ShuffleCardBro{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil

	err = clients.SendToUsers(ids, t, enum.MsgShuffleCard, rs,enum.MsgShuffleCardCode)
	if err != nil {
		return err
	}
	return nil
}

func BankerListHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbroom.UserBankerList{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	ids := rs.Ids
	rs.Ids = nil
	err = clients.SendToUsers(ids, t, enum.MsgBankerList, rs,enum.MsgBankerListCode)
	if err != nil {
		return err
	}
	return nil
}