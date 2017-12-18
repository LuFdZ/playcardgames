package niuniu

import (
	srvniu "playcards/service/niuniu/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var NiuniuEvent = []string{
	srvniu.TopicNiuniuGameResult,
	srvniu.TopicNiuniuBeBanker,
	srvniu.TopicNiuniuSetBet,
	srvniu.TopicNiuniuAllBet,
	srvniu.TopicNiuniuGameReady,
	srvniu.TopicNiuniuGameStart,
	srvniu.TopicNiuniuCountDown,
}

func SubscribeNiuniuMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(NiuniuEvent)
	return nil
}

func UnsubscribeNiuniuMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(NiuniuEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeNiuniuMessage", auth.RightsPlayer,
		SubscribeNiuniuMessage)
	request.RegisterHandler("UnsubscribeNiuniuMessage", auth.RightsPlayer,
		UnsubscribeNiuniuMessage)
}


