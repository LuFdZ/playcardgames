package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/room/handler"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"
	"playcards/utils/sync"

	micro "github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"RoomSrv.CreateRoom": auth.RightsNone,
	"RoomSrv.EnterRoom":  auth.RightsNone,
	"RoomSrv.LeaveRoom":  auth.RightsNone,
	"RoomSrv.SetReady":   auth.RightsNone,
	"RoomSrv.OutReady":   auth.RightsNone,
	"RoomSrv.Heartbeat":  auth.RightsNone,
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
	gt := sync.NewGlobalTimer()
	h := handler.NewHandler(server, gt)
	server.Handle(server.NewHandler(h))

	err := service.Run()
	gt.Stop()
	env.ErrExit(err)
}
