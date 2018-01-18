package fourcard

import (
	srvfour "playcards/service/fourcard/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var FourCardEvent = []string{
	srvfour.TopicFourCardGameStart,
	srvfour.TopicFourCardSetBet,
	srvfour.TopicFourCardGameReady,
	srvfour.TopicFourCardGameSubmitCard,
	srvfour.TopicFourCardGameResult,
}

func SubscribeFourCardMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(FourCardEvent)
	return nil
}

func UnsubscribeFourCardMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(FourCardEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeFourCardMessage", auth.RightsPlayer,
		SubscribeFourCardMessage)
	request.RegisterHandler("UnsubscribeFourCardMessage", auth.RightsPlayer,
		UnsubscribeFourCardMessage)
}
