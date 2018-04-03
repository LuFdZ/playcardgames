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
	"ClubSrv.CreateClub":                   auth.RightsPlayer,
	"ClubSrv.JoinClub":                     auth.RightsPlayer,
	"ClubSrv.LeaveClub":                    auth.RightsPlayer,
	"ClubSrv.GetClub":                      auth.RightsPlayer,
	"ClubSrv.ClubMemberOfferUpClubCoin":    auth.RightsPlayer,
	"ClubSrv.PageClubMemberJournal":        auth.RightsPlayer,
	"ClubSrv.GetClubMemberCoinRank":        auth.RightsPlayer,
	"ClubSrv.UpdateClubMemberStatus":       auth.RightsPlayer,
	"ClubSrv.PageClubRoom":                 auth.RightsPlayer,
	"ClubSrv.PageClubMember":               auth.RightsPlayer,
	"ClubSrv.RemoveClubMember":             auth.RightsPlayer,
	"ClubSrv.SetBlackList":                 auth.RightsPlayer,
	"ClubSrv.UpdateClubExamine":            auth.RightsPlayer,
	"ClubSrv.SetClubRoomFlag":              auth.RightsPlayer,
	"ClubSrv.AddClubMemberClubCoin":        auth.RightsPlayer,
	"ClubSrv.PageClubJournal":              auth.RightsPlayer,
	"ClubSrv.UpdateClubJournal":            auth.RightsPlayer,
	"ClubSrv.GetClubsByMemberID":           auth.RightsPlayer,
	"ClubSrv.CreateVipRoomSetting":         auth.RightsPlayer,
	"ClubSrv.UpdateVipRoomSetting":         auth.RightsPlayer,
	"ClubSrv.GetVipRoomSettingList":        auth.RightsPlayer,
	"ClubSrv.ClubReName":                   auth.RightsPlayer,
	"ClubSrv.ClubDelete":                   auth.RightsPlayer,
	"ClubSrv.GetClubRoomLog":               auth.RightsPlayer,
	"ClubSrv.PageBlackListMember":          auth.RightsPlayer,
	"ClubSrv.CancelBlackList":              auth.RightsPlayer,
	"ClubSrv.PageClubExamineMember":        auth.RightsPlayer,
	"ClubSrv.GetClubByClubID":              auth.RightsPlayer,
	"ClubSrv.ClubRecharge":                 auth.RightsPlayer,
	"ClubSrv.UpdateVipRoomSettingStatus":   auth.RightsPlayer,
	"ClubSrv.UpdateClub":                   auth.RightsPlayer,
	"ClubSrv.PageClubRoomResultList":       auth.RightsPlayer,
	"ClubSrv.PageClubMemberRoomResultList": auth.RightsPlayer,
	"ClubSrv.PageClub":                     auth.RightsClubAdmin,
	"ClubSrv.CreateClubMember":             auth.RightsAdmin,
	"ClubSrv.UpdateClubProxyID":            auth.RightsAdmin,
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
