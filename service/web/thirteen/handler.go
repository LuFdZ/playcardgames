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

	err = clients.SendTo(rs.UserID, t, enum.MsgThireteenGameStart, rs)
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
	//fmt.Printf("ThirteenResultHandler:%+v", rs)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgThireteenGameResult, rs)
	if err != nil {
		return err
	}
	return nil
}

// func ThirteenSurrenderHandler(p broker.Publication) error {
// 	t := p.Topic()
// 	msg := p.Message()
// 	rs := &pbthirteen.SurrenderMessage{}
// 	err := proto.Unmarshal(msg.Body, rs)
// 	if err != nil {
// 		return err
// 	}
// 	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgThireteenGameResult, rs)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func ThirteenReadyHandler(p broker.Publication) error {
	t := p.Topic()
	msg := p.Message()
	rs := &pbthirteen.GameReady{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	//fmt.Printf("ThirteenReadyHandler:%+v", rs)
	err = clients.SendRoomUsers(rs.RoomID, t, enum.MsgThireteenGameReady, rs)
	if err != nil {
		return err
	}
	return nil
}
