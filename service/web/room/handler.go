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
	srvroom.TopicRoomFourCardExist,
	srvroom.TopicRoomTwoCardExist,
	srvroom.TopicRoomRunCardExist,
	srvroom.TopicRoomNotice,
	srvroom.TopicRoomShuffleCardBro,
	srvroom.TopicRoomChat,
	srvroom.TopicBankerList,
	srvroom.TopicUserRestore,
}

func RoomOnlineNotice(c *clients.Client) {
	cacheuser.SetUserOnlineStatus(c.UserID(), 1)
	cacheuser.UpdateUserHeartbeat(c.UserID())
	c.SendMessage("", enum.MsgSubscribeSuccess, nil, enum.MsgSubscribeSuccessCode)

	msg := &pbroom.UserConnection{
		UserID: c.UserID(),
		Status: enum.SocketAline,
	}
	topic.Publish(brok, msg, srvroom.TopicRoomUserConnection)
}

func SubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(RoomEvent)
	return nil
}

func UnsubscribeRoomMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RoomEvent)
	return nil
}

func CloseCallbackHandler(c *clients.Client) {
	//cacheuser.SetUserOnlineStatus(c.UserID(), 0)
	cs := clients.GetClientByUserID(c.UserID())
	if len(cs) == 0 {
		msg := &pbroom.UserConnection{
			UserID: c.UserID(),
			Status: enum.SocketClose,
		}
		topic.Publish(brok, msg, srvroom.TopicRoomUserConnection)
	}
}

func ReceiveTestLogMessage(c *clients.Client, req *request.Request) error {
	var RecType string
	err := json.Unmarshal(req.Args, &RecType)
	if err != nil {
		RecType = "Err"
	}
	log.Debug("ReceiveTestLog:%d|%v", c.UserID(), RecType)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeRoomMessage", auth.RightsPlayer,
		SubscribeRoomMessage)
	request.RegisterHandler("UnSubscribeRoomMessage", auth.RightsPlayer,
		UnsubscribeRoomMessage)
	request.RegisterHandler("ReceiveTestLog", auth.RightsPlayer,
		ReceiveTestLogMessage)
	request.RegisterCloseHandler(CloseCallbackHandler)
}
