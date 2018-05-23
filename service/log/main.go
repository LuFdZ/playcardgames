package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/log/handler"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"
	"playcards/utils/sync"
	"playcards/utils/auth"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"LogSrv.ClientReportException": auth.RightsNone,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.LogServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.LogServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()
	gt := sync.NewGlobalTimer()
	server := service.Server()
	h := handler.NewHandler(service, gt)
	server.Handle(server.NewHandler(h))

	err := service.Run()
	env.ErrExit(err)
}
