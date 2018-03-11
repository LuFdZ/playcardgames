package main

import (
	"playcards/service/api/enum"
	"playcards/service/club/handler"
	envinit "playcards/service/init"
	"playcards/utils/auth"
	gcf "playcards/utils/config"
	"playcards/utils/env"
	"playcards/utils/log"

	"github.com/micro/go-micro"
)

var FuncRights = map[string]int32{
	"ClubSrv.CreateClub":                auth.RightsPlayer,
	"ClubSrv.UpdateClub":                auth.RightsClubAdmin,
	"ClubSrv.CreateClubMember":          auth.RightsClubAdmin,
	"ClubSrv.JoinClub":                  auth.RightsPlayer,
	"ClubSrv.LeaveClub":                 auth.RightsPlayer,
	"ClubSrv.GetClub":                   auth.RightsPlayer,
	"ClubSrv.ClubMemberOfferUpClubCoin": auth.RightsPlayer,
	"ClubSrv.PageClubMemberJournal":     auth.RightsPlayer,
	"ClubSrv.GetClubMemberCoinRank":     auth.RightsPlayer,
	"ClubSrv.UpdateClubMemberStatus":    auth.RightsPlayer,
	"ClubSrv.PageClubRoom":              auth.RightsClubAdmin,
	"ClubSrv.PageClub":                  auth.RightsClubAdmin,
	"ClubSrv.PageClubMember":            auth.RightsClubAdmin,
	"ClubSrv.ClubRecharge":              auth.RightsAdmin,
	"ClubSrv.RemoveClubMember":          auth.RightsAdmin,
	"ClubSrv.SetBlackList":              auth.RightsAdmin,
	"ClubSrv.UpdateClubExamine":         auth.RightsAdmin,
	"ClubSrv.SetClubRoomFlag":           auth.RightsAdmin,
	"ClubSrv.AddClubMemberClubCoin":     auth.RightsAdmin,
	"ClubSrv.PageClubJournal":           auth.RightsAdmin,
	"ClubSrv.UpdateClubJournal":         auth.RightsAdmin,
}

func main() {
	envinit.Init()
	log.Info("start %s", enum.ClubServiceName)

	ttl, interval := gcf.RegisterTTL()

	service := micro.NewService(
		micro.Name(enum.ClubServiceName),
		micro.Version(enum.VERSION),
		micro.RegisterTTL(ttl),
		micro.RegisterInterval(interval),
		micro.WrapHandler(auth.ServerAuthWrapper(FuncRights)),
	)
	service.Init()
	server := service.Server()
	h := handler.NewHandler(server)
	server.Handle(server.NewHandler(h))

	err := service.Run()
	env.ErrExit(err)
}
