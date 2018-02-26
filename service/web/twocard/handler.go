package twocard

import (
	srvtow "playcards/service/twocard/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var TwoCardEvent = []string{
	srvtow.TopicTwoCardGameStart,
	srvtow.TopicTwoCardSetBet,
	srvtow.TopicTwoCardGameReady,
	srvtow.TopicTwoCardGameSubmitCard,
	srvtow.TopicTwoCardGameResult,
}

func SubscribeTwoCardMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(TwoCardEvent)
	return nil
}

func UnsubscribeTwoCardMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(TwoCardEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeTwoCardMessage", auth.RightsPlayer,
		SubscribeTwoCardMessage)
	request.RegisterHandler("UnsubscribeTwoCardMessage", auth.RightsPlayer,
		UnsubscribeTwoCardMessage)
}
