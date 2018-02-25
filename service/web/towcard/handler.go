package fourcard

import (
	srvtow "playcards/service/towcard/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var TowCardEvent = []string{
	srvtow.TopicTowCardGameStart,
	srvtow.TopicTowCardSetBet,
	srvtow.TopicTowCardGameReady,
	srvtow.TopicTowCardGameSubmitCard,
	srvtow.TopicTowCardGameResult,
}

func SubscribeTowCardMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(TowCardEvent)
	return nil
}

func UnsubscribeTowCardMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(TowCardEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeTowCardMessage", auth.RightsPlayer,
		SubscribeTowCardMessage)
	request.RegisterHandler("UnsubscribeTowCardMessage", auth.RightsPlayer,
		UnsubscribeTowCardMessage)
}
