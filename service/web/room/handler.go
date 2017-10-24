package room

import (
	"encoding/json"
	cacheroom "playcards/model/room/cache"
	"playcards/model/room/errors"
	pbroom "playcards/proto/room"
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	enum "playcards/service/web/enum"
	"playcards/service/web/request"
	"playcards/utils/auth"
	gctx "playcards/utils/context"
	"playcards/utils/log"
)

var RoomEvent = []string{
	srvroom.TopicRoomStatusChange,
	srvroom.TopicRoomReady,
	srvroom.TopicRoomUnReady,
	srvroom.TopicRoomJoin,
	srvroom.TopicRoomUnJoin,
	srvroom.TopicRoomResult,
	srvroom.TopicRoomGiveup,
	srvroom.TopicRoomShock,
	srvroom.TopicRoomUserConnection,
	srvroom.TopicRoomRenewal,
	srvroom.TopicRoomVoiceChat,
}

func SubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	var pwd string
	err := json.Unmarshal(req.Args, &pwd)
	if err != nil {
		return err
	}
	room, err := cacheroom.GetRoom(pwd)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.ErrRoomNotExisted
	}
	err = cacheroom.UpdateRoomUserSocektStatus(c.User().UserID, enum.SocketAline, 0)
	if err != nil {
		return err
	}
	c.User().RoomID = room.RoomID
	c.Subscribe(RoomEvent)
	return nil
}

func UnsubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RoomEvent)
	return nil
}

func ClinetHearbeatMessage(c *clients.Client,req *request.Request)error{
	var msg byte
	msg=1
	c.SendMessage(enum.MsgHearbeat,enum.MsgHearbeat,msg)
	return nil
}

func HeartbeatCallback(c *clients.Client) {
	log.Debug("room heartbeat %v", c)
	ctx := gctx.NewContext(c.Token())
	rpc.Heartbeat(ctx, &pbroom.HeartbeatRequest{})
}

func init() {
	request.RegisterHandler("SubscribeRoomMessage", auth.RightsPlayer,
		SubscribeRoomMessage)
	request.RegisterHandler("UnSubscribeRoomMessage", auth.RightsPlayer,
		UnsubscribeRoomMessage)
	request.RegisterHandler("ClinetHearbeat", auth.RightsPlayer,
		ClinetHearbeatMessage)
	request.RegisterHeartbeatHandler(HeartbeatCallback)
}
