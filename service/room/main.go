package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/room/handler"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	micro "github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"RoomSrv.CreateRoom": auth.RightsNone,
	"RoomSrv.JoinRoom":   auth.RightsNone,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.RoomServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.RoomServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	server.Handle(server.NewHandler(&handler.RoomSrv{}))

	err := service.Run()
	env.ErrExit(err)
}
