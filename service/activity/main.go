package main

import (
	"playcards/service/activity/handler"
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"playcards/utils/auth"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"ActivitySrv.AddConfig":    auth.RightsActivityEdit,
	"ActivitySrv.DeleteConfig": auth.RightsActivityEdit,
	"ActivitySrv.ListConfig":   auth.RightsActivityEdit,
	"ActivitySrv.UpdateConfig": auth.RightsActivityEdit,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.ActivityServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.ActivityServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	server.Handle(server.NewHandler(&handler.ActivitySrv{}))

	err := service.Run()
	env.ErrExit(err)
}
