package main

import (
	"playcards/service/api/enum"
	"playcards/service/common/handler"
	envinit "playcards/service/init"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"CommonSrv.CreateBlackList": auth.RightsPlayer,
	"CommonSrv.CreateExamine":   auth.RightsPlayer,
	"CommonSrv.CancelBlackList": auth.RightsPlayer,
	"CommonSrv.UpdateExamine":   auth.RightsPlayer,
	"CommonSrv.PageBlackList":   auth.RightsPlayer,
	"CommonSrv.PageExamine":     auth.RightsPlayer,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.CommonServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.CommonServiceName),
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
