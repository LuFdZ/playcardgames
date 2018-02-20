package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/mail/handler"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"
	"playcards/utils/sync"

	micro "github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"MailSrv.SendMail":                    auth.RightsPlayer,
	"MailSrv.CreatePlayerMails":           auth.RightsPlayer,
	"MailSrv.ReadMail":                    auth.RightsPlayer,
	"MailSrv.GetMailItems":                auth.RightsPlayer,
	"MailSrv.PagePlayerMail":              auth.RightsPlayer,
	"MailSrv.PageMailInfo":                auth.RightsPlayer,
	"MailSrv.PageMailSendLog":             auth.RightsPlayer,
	"MailSrv.PageAllPlayerMail":           auth.RightsPlayer,
	"MailSrv.GetAllMailItems":             auth.RightsPlayer,
	"MailSrv.SendSysMail":                 auth.RightsAdmin,
	"MailSrv.CreateMailInfo":              auth.RightsAdmin,
	"MailSrv.UpdateMailInfo":              auth.RightsAdmin,
	"MailSrv.RefreshAllMailInfoFromDB":    auth.RightsAdmin,
	"MailSrv.RefreshAllMailSendLogFromDB": auth.RightsAdmin,
	"MailSrv.RefreshAllPlayerMailFromDB":  auth.RightsAdmin,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.MailServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.MailServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	gt := sync.NewGlobalTimer()

	var err error
	h := handler.NewHandler(server, gt)
	server.Handle(server.NewHandler(h))

	err = service.Run()
	gt.Stop()
	env.ErrExit(err)
}
