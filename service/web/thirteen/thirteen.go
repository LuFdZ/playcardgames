package thirteen

import (
	srvthirteen "playcards/service/thirteen/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var ThirteenEvent = []string{
	srvthirteen.TopicThirteenGameResult,
	srvthirteen.TopicThirteenSurrender,
}

func SubscribeThirteenMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(ThirteenEvent)
	return nil
}

func UnsubscribeThirteenMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(ThirteenEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeThirteenMessage", auth.RightsPlayer,
		SubscribeThirteenMessage)
	request.RegisterHandler("UnsubscribeThirteenMessage", auth.RightsPlayer,
		UnsubscribeThirteenMessage)
}
