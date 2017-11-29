package room

import (
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
	gctx "playcards/utils/context"
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
	srvroom.TopicRoomExist,
	srvroom.TopicRoomNotice,
}

func SubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(RoomEvent)
	c.SendNewClientBackMessage("SubscribeSuccess")
	return nil
}

func UnsubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RoomEvent)
	return nil
}

func ClinetHearbeatMessage(c *clients.Client, req *request.Request) error {
	c.SendHearbeatMessage()
	return nil
}

func HeartbeatCallback(c *clients.Client) {
	//log.Debug("room heartbeat %v", c)
	ctx := gctx.NewContext(c.Token())
	rpc.Heartbeat(ctx, &pbroom.HeartbeatRequest{})
}

func init() {
	request.RegisterHandler("SubscribeRoomMessage", auth.RightsPlayer,
		SubscribeRoomMessage)
	request.RegisterHandler("UnSubscribeRoomMessage", auth.RightsPlayer,
		UnsubscribeRoomMessage)
	request.RegisterHandler("ClientHeartbeat", auth.RightsPlayer,
		ClinetHearbeatMessage)
	request.RegisterHeartbeatHandler(HeartbeatCallback)
}
