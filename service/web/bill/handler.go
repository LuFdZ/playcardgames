package bill

import (
	srvbill "playcards/service/bill/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var BillEvent = []string{
	srvbill.TopicBillChange,
}

func SubscribeBillMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(BillEvent)
	return nil
}

func UnsubscribeBillMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(BillEvent)
	return nil
}



func init() {
	request.RegisterHandler("SubscribeBillMessage", auth.RightsPlayer,
		SubscribeBillMessage)
	request.RegisterHandler("UnsubscribeBillMessage", auth.RightsPlayer,
		UnsubscribeBillMessage)

}
