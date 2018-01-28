package mail

import (
	srvmail "playcards/service/mail/handler"
	"playcards/service/web/clients"
	"playcards/service/web/request"
	"playcards/utils/auth"
)

var MailEvent = []string{
	srvmail.TopicMailNotice,
}

func SubscribeMailMessage(c *clients.Client, req *request.Request) error {
	c.Subscribe(MailEvent)
	return nil
}

func UnsubscribeMailMessage(c *clients.Client, req *request.Request) error {
	c.Unsubscribe(MailEvent)
	return nil
}

func init() {
	request.RegisterHandler("SubscribeMailMessage", auth.RightsPlayer,
		SubscribeMailMessage)
	request.RegisterHandler("UnsubscribeMailMessage", auth.RightsPlayer,
		UnsubscribeMailMessage)
}
