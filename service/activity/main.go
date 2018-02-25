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
	"ActivitySrv.Share":          auth.RightsPlayer,
	"ActivitySrv.Invite":         auth.RightsPlayer,
	"ActivitySrv.InviteUserInfo": auth.RightsPlayer,
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
	h := handler.NewHandler(server)
	server.Handle(server.NewHandler(h))

	err := service.Run()
	env.ErrExit(err)
}
