package main

import (
"playcards/service/api/enum"
"playcards/service/init"
"playcards/service/goldroom/handler"
"playcards/utils/auth"
gcf "playcards/utils/config"
"playcards/utils/env"
"playcards/utils/log"
"playcards/utils/sync"

"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"GoldRoomSrv.EnterRoom":             auth.RightsPlayer,
	"GoldRoomSrv.LeaveRoom":             auth.RightsPlayer,
	"GoldRoomSrv.SetReady":              auth.RightsPlayer,
	"GoldRoomSrv.GiveUpGame":            auth.RightsPlayer,
	"GoldRoomSrv.GiveUpVote":            auth.RightsPlayer,
	"GoldRoomSrv.Renewal":               auth.RightsPlayer,
	"GoldRoomSrv.RoomResultList":        auth.RightsPlayer,
	"GoldRoomSrv.Shock":                 auth.RightsPlayer,
	"GoldRoomSrv.VoiceChat":             auth.RightsPlayer,
	"GoldRoomSrv.GetAgentRoomList":      auth.RightsPlayer,
	"GoldRoomSrv.DeleteAgentRoomRecord": auth.RightsPlayer,
	"GoldRoomSrv.DisbandAgentRoom":      auth.RightsPlayer,
	"GoldRoomSrv.GetRoomUserLocation":   auth.RightsPlayer,
	"GoldRoomSrv.CreateClubRoom":        auth.RightsPlayer,
	"GoldRoomSrv.CreateFeedback":        auth.RightsPlayer,
	"GoldRoomSrv.GetRoomRecovery":       auth.RightsPlayer,
	"GoldRoomSrv.GameStart":             auth.RightsPlayer,
	"GoldRoomSrv.ShuffleCard":           auth.RightsPlayer,
	"GoldRoomSrv.RoomChat":              auth.RightsPlayer,
	"GoldRoomSrv.PageSpecialGameList":   auth.RightsAdmin,
	"GoldRoomSrv.PageRoomList":          auth.RightsNoticeAdmin,
	"GoldRoomSrv.PageFeedbackList":      auth.RightsRoomAdmin,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.GoldRoomServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.GoldRoomServiceName),
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
