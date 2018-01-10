package doudizhu

import (
srvddz "playcards/service/doudizhu/handler"
"playcards/service/web/clients"
"playcards/service/web/request"
"playcards/utils/auth"
)

var DoudizhuEvent = []string{
	srvddz.TopicDDZGameStart,
	srvddz.TopicDDZBeBanker,
	srvddz.TopicDDZSubmitCard,
	srvddz.TopicDDZGameResult,
}

func SubscribeDoudizhuMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(DoudizhuEvent)
	return nil
}

func UnsubscribeDoudizhuMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(DoudizhuEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeDouDiZhuMessage", auth.RightsPlayer,
		SubscribeDoudizhuMessage)
	request.RegisterHandler("UnsubscribeDouDiZhuMessage", auth.RightsPlayer,
		UnsubscribeDoudizhuMessage)
}
