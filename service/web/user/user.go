package user

import (
	apienum "playcards/service/api/enum"
	pbuser "playcards/proto/user"
	srvuser "playcards/service/user/handler"
	"playcards/service/web/clients"
	"playcards/utils/subscribe"
	"playcards/utils/topic"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/protobuf/proto"
	"playcards/utils/log"
)

var rpc pbuser.UserSrvClient

var (
	brok broker.Broker
)

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAllUserMessage(brk); err != nil {
		return err
	}

	rpc = pbuser.NewUserSrvClient(
		apienum.UserServiceName,
		client.DefaultClient,
	)
	return nil
}

func SubscribeAllUserMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvuser.TopicHeartbeatTimeout),
		HeartbeatTimeoutHandler,
	)
	return nil
}

func HeartbeatTimeoutHandler(p broker.Publication) error {
	msg := p.Message()
	rs := &pbuser.User{}
	err := proto.Unmarshal(msg.Body, rs)
	if err != nil {
		return err
	}
	cs := clients.GetClientByUserID(rs.UserID)
	for _, c := range cs {
		c.Close()
		log.Debug("heartbeat timeout handler:%d\n", rs.UserID)
	}
	return nil
}
