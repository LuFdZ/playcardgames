package main

import (
	"playcards/service/api/enum"
	envinit "playcards/service/init"
	"playcards/service/web/handler"
	"playcards/utils/config"
	"playcards/utils/log"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	"golang.org/x/net/websocket"
)

func main() {
	envinit.Init()
	log.Info("start %s", enum.WebServiceName)
	address := config.WebHost()

	service := web.NewService(
		web.Name(enum.WebServiceName),
		web.Address(address),
	)
	service.Init()

	handler := handler.NewWebHandler(client.DefaultClient)
	service.Handle("/stream", websocket.Handler(handler.Subscribe))
	service.Run()
	handler.Stop()
}
