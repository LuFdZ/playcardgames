package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/notice/handler"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	micro "github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"NoticeSrv.GetNotice":      auth.RightsPlayer,
	"NoticeSrv.AllNotice":      auth.RightsNoticeView,
	"NoticeSrv.CreateNotice":   auth.RightsNoticeAdmin,
	"NoticeSrv.UpdateNotice":   auth.RightsNoticeAdmin,
	"NoticeSrv.PageNoticeList": auth.RightsNoticeAdmin,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.NoticeServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.NoticeServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	server.Handle(server.NewHandler(&handler.NoticeSrv{}))

	err := service.Run()
	env.ErrExit(err)
}
