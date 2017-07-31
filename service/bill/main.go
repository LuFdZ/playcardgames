package main

import (
	"playcards/service/api/enum"
	"playcards/service/bill/handler"
	envinit "playcards/service/init"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"BillSrv.GetBalance":  auth.RightsPlayer | auth.RightsBillView,
	"BillSrv.GainBalance": auth.RightsBillEdit,
	"BillSrv.Recharge":    auth.RightsBillEdit,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.BillServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.BillServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	server.Handle(server.NewHandler(&handler.BillSrv{}))

	err := service.Run()
	env.ErrExit(err)
}
