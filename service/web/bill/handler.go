package bill

import (
	cacheuser "playcards/model/user/cache"
	srvbill "playcards/service/bill/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var BillEvent = []string{
	srvbill.TopicBillChange,
}

func SubscribeBillMessage(c *clients.Client, req *request.Request) error {
	cacheuser.SetUserOnlineStatus(c.UserID(), 1)
	c.Subscribe(BillEvent)
	return nil
}

func UnsubscribeBillMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(BillEvent)
	return nil
}

func CloseCallbackHandler(c *clients.Client) {
	cacheuser.SetUserOnlineStatus(c.UserID(), 0)
}

func init() {
	request.RegisterHandler("SubscribeBillMessage", auth.RightsPlayer,
		SubscribeBillMessage)
	request.RegisterHandler("UnsubscribeBillMessage", auth.RightsPlayer,
		UnsubscribeBillMessage)
	request.RegisterCloseHandler(CloseCallbackHandler)
}
