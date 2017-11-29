package club

import (
	cacheuser "playcards/model/user/cache"
	pbclub "playcards/proto/club"
	srvclub "playcards/service/club/handler"
	srvroom "playcards/service/room/handler"
	"playcards/service/web/clients"
	enumweb "playcards/service/web/enum"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var ClubEvent = []string{
	srvclub.TopicClubMemberJoin,
	srvclub.TopicClubMemberLeave,
	srvclub.TopicClubInfo,
	srvclub.TopicClubOnlineStatus,
	srvroom.TopicClubRoomCreate,
	srvroom.TopicClubRoomJoin,
	srvroom.TopicClubRoomUnJoin,
	srvroom.TopicClubRoomFinish,
}

func ClubOnlineNotice(c *clients.Client) {
	ClubUserOlineChange(c.UserID(), c.User().ClubID, enumweb.SocketAline)
}

func SubscribeClubMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(ClubEvent)
	_, user := cacheuser.GetUserByID(c.UserID())
	mo := &pbclub.ClubMemberOnline{}
	if user != nil {
		mo.UserID = user.UserID
		mo.ClubID = user.ClubID
	}

	c.SendMessage("", "ClubSubscribeSuccess", mo)
	//ClubUserOlineChange(c.UserID(), c.User().ClubID, enumweb.SocketAline)
	return nil
}

func UnsubscribeClubMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(ClubEvent)
	return nil
}

func CloseCallbackHandler(c *clients.Client) {
	ClubUserOlineChange(c.UserID(), c.User().ClubID, enumweb.SocketClose)
}

func ClubUserOlineChange(uid int32, clubid int32, status int32) {
	if clubid == 0 {
		return
	}
	mo := &pbclub.ClubMemberOnline{
		UserID: uid,
		Status: status,
		ClubID: clubid,
	}
	ClubOnlineStatus(srvclub.TopicClubOnlineStatus, mo)
}

func init() {
	request.RegisterHandler("SubscribeClubMessage", auth.RightsPlayer,
		SubscribeClubMessage)
	request.RegisterHandler("UnSubscribeClubMessage", auth.RightsPlayer,
		UnsubscribeClubMessage)
	request.RegisterCloseHandler(CloseCallbackHandler)
}
