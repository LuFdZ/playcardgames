package thirteen

import (
	pbthirteen "playcards/proto/thirteen"
	srvthirteen "playcards/service/thirteen/handler"
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
	if err := SubscribeAllThirteenMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAllThirteenMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
		TopicThirteenGameResult),
		ThirteenGameResultHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
		TopicThirteenSurrender),
		ThirteenSurrenderHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvthirteen.
		TopicThirteenGameReady),
		ThirteenReadyHandler,
	)
	return nil
}

func ThirteenGameResultHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgThireteenGameResult, rs)
	if err != nil {
		return err
	}
	return nil
}

func ThirteenSurrenderHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GameResult{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgThireteenGameResult, rs)
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
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgRoomUnReady, rs)
	if err != nil {
		return err
	}
	return nil
}
