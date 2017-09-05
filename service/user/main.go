package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/user/handler"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"playcards/utils/auth"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"UserSrv.Register":       auth.RightsNone,
	"UserSrv.Login":          auth.RightsNone,
	"UserSrv.GetProperty":    auth.RightsPlayer | auth.RightsUserView,
	"UserSrv.UserInfo":       auth.RightsUserView,
	"UserSrv.PageUserList":   auth.RightsUserView,
	"UserSrv.GetToken":       auth.RightsUserView,
	"UserSrv.CheckUser":      auth.RightsUserView,
	"UserSrv.GetUser":        auth.RightsUserView,
	"UserSrv.UpdateUser":     auth.RightsUserEdit,
	"UserSrv.UpdateProperty": auth.RightsUserEdit,
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
	server.Handle(server.NewHandler(&handler.UserSrv{}))

	err := service.Run()
	env.ErrExit(err)
}
