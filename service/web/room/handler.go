package room

import (
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var RoomEvent = []string{
	srvroom.TopicRoomStatusChange,
	srvroom.TopicRoomReady,
	srvroom.TopicRoomUnReady,
	srvroom.TopicRoomJoin,
	srvroom.TopicRoomUnJoin,
}

func SubscribeRoomMessageJoin(c *clients.Client, req *request.Request) error {
	c.Subscribe(RoomEvent)
	return nil
}

func UnsubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RoomEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeRoomMessage", auth.RightsPlayer,
		SubscribeRoomMessage)
	request.RegisterHandler("UnsubscribeRoomMessage", auth.RightsPlayer,
		UnsubscribeRoomMessage)
}
