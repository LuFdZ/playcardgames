package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/towcard/handler"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"
	"playcards/utils/sync"
	"time"
	"fmt"

	micro "github.com/micro/go-micro"
	"github.com/yuin/gopher-lua"
)

var FuncRights = map[string]int32{
	"TowCardSrv.SubmitCard":      auth.RightsPlayer,
	"TowCardSrv.SetBet":          auth.RightsPlayer,
	"TowCardSrv.GameResultList":  auth.RightsPlayer,
	"TowCardSrv.TowCardRecovery": auth.RightsPlayer,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.TowCardServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.TowCardServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()

	server := service.Server()
	gt := sync.NewGlobalTimer()
	l := lua.NewState()
	defer l.Close()
	var err error
	filePath := env.GetCurrentDirectory() + "/lua/towcardlua/Logic.lua"
	if err = l.DoFile(filePath); err != nil {
		log.Err("tow card logic do file %+v", err)
	}

	ostime := time.Now().UnixNano()
	if err = l.DoString(fmt.Sprintf("return G_Init(%d)", ostime)); err != nil {
		log.Err("tow card G_Init error %+v", err)
	}

	h := handler.NewHandler(server, gt, l)
	server.Handle(server.NewHandler(h))

	err = service.Run()
	gt.Stop()
	env.ErrExit(err)
}