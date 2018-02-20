package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/user/handler"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"
	"playcards/utils/sync"
	"playcards/utils/auth"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"UserSrv.Register":               auth.RightsNone,
	"UserSrv.Login":                  auth.RightsNone,
	"UserSrv.WXLogin":                auth.RightsNone,
	"UserSrv.GetProperty":            auth.RightsPlayer | auth.RightsUserView,
	"UserSrv.SetLocation":            auth.RightsPlayer,
	"UserSrv.Heartbeat":              auth.RightsPlayer,
	"UserSrv.UserInfo":               auth.RightsUserView,
	"UserSrv.PageUserList":           auth.RightsUserView,
	"UserSrv.GetToken":               auth.RightsUserView,
	"UserSrv.CheckUser":              auth.RightsUserView,
	"UserSrv.GetUser":                auth.RightsUserView,
	"UserSrv.DayActiveUserList":      auth.RightsUserView,
	"UserSrv.GetUserCount":           auth.RightsUserView,
	"UserSrv.RefreshUserCount":       auth.RightsUserView,
	"UserSrv.UpdateUser":             auth.RightsUserEdit,
	"UserSrv.UpdateProperty":         auth.RightsUserEdit,
	"UserSrv.RefreshAllRobotsFromDB": auth.RightsAdmin,
	"UserSrv.RegisterRobot": auth.RightsAdmin,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.UserServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.UserServiceName),
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

	//server := service.Server()
	//server.Handle(server.NewHandler(&handler.UserSrv{}))

	err := service.Run()
	gt.Stop()
	env.ErrExit(err)
}
