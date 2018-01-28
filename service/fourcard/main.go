package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/fourcard/handler"
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
	"FourCardSrv.SubmitCard":       auth.RightsPlayer,
	"FourCardSrv.SetBet":           auth.RightsPlayer,
	"FourCardSrv.GameResultList":   auth.RightsPlayer,
	"FourCardSrv.FourCardRecovery": auth.RightsPlayer,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.FourCardServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.FourCardServiceName),
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
	filePath := env.GetCurrentDirectory() + "/lua/fourcardlua/Logic.lua"
	if err = l.DoFile(filePath); err != nil {
		log.Err("four card logic do file %+v", err)
	}

	ostime := time.Now().UnixNano()
	if err = l.DoString(fmt.Sprintf("return G_Init(%d)", ostime)); err != nil {
		log.Err("four card G_Init error %+v", err)
	}

	h := handler.NewHandler(server, gt, l)
	server.Handle(server.NewHandler(h))

	err = service.Run()
	gt.Stop()
	env.ErrExit(err)
}
