package main

import (
	"playcards/service/api/enum"
	"playcards/service/config/handler"
	envinit "playcards/service/init"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"ConfigSrv.UpdateConfig":            auth.RightsConfigEdit,
	"ConfigSrv.CreateConfig":            auth.RightsConfigEdit,
	"ConfigSrv.GetConfigs":              auth.RightsPlayer,
	"ConfigSrv.GetConfigsBeforeLogin":   auth.RightsNone,
	"ConfigSrv.PageConfigs":             auth.RightsConfigEdit,
	"ConfigSrv.RefreshAllConfigsFromDB": auth.RightsConfigEdit,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.ConfigServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.ConfigServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	h := handler.NewHandler()
	server.Handle(server.NewHandler(h))

	err := service.Run()
	env.ErrExit(err)
}
