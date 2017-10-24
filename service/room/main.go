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
	"RoomSrv.Heartbeat":        auth.RightsNone,
	"RoomSrv.CreateRoom":       auth.RightsPlayer,
	"RoomSrv.EnterRoom":        auth.RightsPlayer,
	"RoomSrv.LeaveRoom":        auth.RightsPlayer,
	"RoomSrv.SetReady":         auth.RightsPlayer,
	"RoomSrv.GiveUpGame":       auth.RightsPlayer,
	"RoomSrv.GiveUpVote":       auth.RightsPlayer,
	"RoomSrv.Renewal":          auth.RightsPlayer,
	"RoomSrv.RoomResultList":   auth.RightsPlayer,
	"RoomSrv.CheckRoomExist":   auth.RightsPlayer,
	"RoomSrv.Shock":            auth.RightsPlayer,
	"RoomSrv.VoiceChat":        auth.RightsPlayer,
	"RoomSrv.PageFeedbackList": auth.RightsRoomAdmin,
	"RoomSrv.CreateFeedback":   auth.RightsRoomAdmin,
	"RoomSrv.TestClean":        auth.RightsNone,
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
