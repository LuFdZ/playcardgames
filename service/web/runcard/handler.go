package twocard

import (
	srvrun "playcards/service/runcard/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var RunCardEvent = []string{
	srvrun.TopicRunCardGameStart,
	srvrun.TopicRunCardGameSubmitCard,
	srvrun.TopicRunCardGameResult,
}

func SubscribeRunCardMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(RunCardEvent)
	return nil
}

func UnsubscribeRunCardMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(RunCardEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeRunCardMessage", auth.RightsPlayer,
		SubscribeRunCardMessage)
	request.RegisterHandler("UnsubscribeRunCardMessage", auth.RightsPlayer,
		UnsubscribeRunCardMessage)
}
