package room

import (
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	cacheuser "playcards/model/user/cache"
	"playcards/utils/topic"
	"playcards/service/web/enum"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
	"encoding/json"
	"playcards/utils/log"
	gctx "playcards/utils/context"
	"time"
)

var RoomEvent = []string{
	srvroom.TopicRoomCreate,
	srvroom.TopicRoomReady,
	srvroom.TopicRoomJoin,
	srvroom.TopicRoomUnJoin,
	srvroom.TopicRoomResult,
	srvroom.TopicRoomGiveup,
	srvroom.TopicRoomShock,
	srvroom.TopicRoomUserConnection,
	srvroom.TopicRoomRenewal,
	srvroom.TopicRoomVoiceChat,
	srvroom.TopicRoomThirteenExist,
	srvroom.TopicRoomNiuniuExist,
	srvroom.TopicRoomDoudizhuExist,
	srvroom.TopicRoomNotice,
}

func SubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	cacheuser.SetUserOnlineStatus(c.UserID(), 1)
	c.Subscribe(RoomEvent)
	c.SendMessage("",enum.MsgSubscribeSuccess,nil)

	msg := &pbroom.UserConnection{
		UserID: c.UserID(),
		Status: enum.SocketAline,
	}
	topic.Publish(brok, msg, srvroom.TopicRoomUserConnection)
	return nil
}

func UnsubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RoomEvent)
	return nil
}

func ClinetHearbeatMessage(c *clients.Client, req *request.Request) error {
	var ctime  int64 = 0

	err := json.Unmarshal(req.Args, &ctime)
	if err != nil {
		ctime = 0
	}
	msg := &pbroom.HeartBeat{}
	if ctime != 0 {
		msg.Stime = time.Now().Unix()
		msg.Ctime = ctime
	}
	c.SendMessage("", enum.MsgHeartbeat, msg)
	return nil
}

func HeartbeatCallback(c *clients.Client) {
	//log.Debug("room heartbeat %v", c)
	ctx := gctx.NewContext(c.Token())
	rpc.Heartbeat(ctx, &pbroom.HeartbeatRequest{})
}

func CloseCallbackHandler(c *clients.Client) {
	cacheuser.SetUserOnlineStatus(c.UserID(), 0)
	msg := &pbroom.UserConnection{
		UserID: c.UserID(),
		Status: enum.SocketClose,
	}
	topic.Publish(brok, msg, srvroom.TopicRoomUserConnection)
}

func ReceiveTestLogMessage(c *clients.Client, req *request.Request) error {
	var RecType  string
	err := json.Unmarshal(req.Args, &RecType)
	if err != nil {
		RecType = "Err"
	}
	log.Debug("ReceiveTestLog:%d|%v",c.UserID(),RecType)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeRoomMessage", auth.RightsPlayer,
		SubscribeRoomMessage)
	request.RegisterHandler("UnSubscribeRoomMessage", auth.RightsPlayer,
		UnsubscribeRoomMessage)
	request.RegisterHandler("ClientHeartbeat", auth.RightsPlayer,
		ClinetHearbeatMessage)
	request.RegisterHandler("ReceiveTestLog", auth.RightsPlayer,
		ReceiveTestLogMessage)
	request.RegisterCloseHandler(CloseCallbackHandler)
	request.RegisterHeartbeatHandler(HeartbeatCallback)
}
