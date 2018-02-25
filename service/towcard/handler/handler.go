package handler

import (
	"context"
	"fmt"
	"playcards/model/towcard"
	mdgame "playcards/model/towcard/mod"
	enumgame "playcards/model/towcard/enum"
	"playcards/model/room"
	enumroom "playcards/model/room/enum"
	pbtow "playcards/proto/towcard"
	"playcards/utils/auth"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/yuin/gopher-lua"
)

type TowCardSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *TowCardSrv {
	tcs := &TowCardSrv{
		server: s,
		broker: s.Options().Broker,
	}
	tcs.update(gt, gl)
	return tcs
}

func (tcs *TowCardSrv) update(gt *gsync.GlobalTimer, gl *lua.LState) {
	lock := "playcards.towcard.update.lock"
	f := func() error {
		tcs.count ++
		newGames := towcard.CreateGame()
		if newGames != nil {
			for _, game := range newGames {
				for _, userResult := range game.GameResult.List {
					msg := &pbtow.GameStart{
						GameID:     game.GameID,
						UserID:     userResult.UserID,
						BankerID:   game.BankerID,
						RoomStatus: enumroom.RoomStatusStarted,
						CardList:   userResult.CardList,
						GameStatus: game.Status,
						CountDown: &pbtow.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SetBetTime,
						},
					}
					topic.Publish(tcs.broker, msg, TopicTowCardGameStart)
				}
			}
		}
		updateGames := towcard.UpdateGame(gl)
		if updateGames != nil {
			for _, game := range updateGames {
				if game.Status == enumgame.GameStatusOrdered {
					msg := game.ToProto()
					for _, ui := range game.GameResult.List {
						var uis []*pbtow.UserInfo
						for _, uiSub := range game.GameResult.List {
							mdui := &pbtow.UserInfo{
								UserID: uiSub.UserID,
								Bet:    uiSub.Bet,
								Role:   uiSub.Role,
								Status: uiSub.Status,
							}
							if ui.UserID == uiSub.UserID {
								mdui.CardList = uiSub.CardList
							}
							uis = append(uis, mdui)
						}
						msg.UserID = ui.UserID
						msg.List = uis
						topic.Publish(tcs.broker, msg, TopicTowCardGameReady)
					}
				} else if game.Status == enumgame.GameStatusDone {
					msg := game.ToProto()
					bro := &pbtow.GameResultBro{
						Context: msg,
						Ids:     game.Ids,
					}
					topic.Publish(tcs.broker, bro, TopicTowCardGameResult)
				}
			}
		}
		if tcs.count == 3 {
			err := towcard.CleanGame()
			if err != nil {
				log.Err("four card clean give up game loop err:%v", err)
			}
			tcs.count = 0
		}

		return nil
	}
	gt.Register(lock, time.Millisecond*enumgame.LoopTime, f)
}

func (fcs *TowCardSrv) SetBet(ctx context.Context, req *pbtow.SetBetRequest,
	rsp *pbtow.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbtow.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	f := func() error {
		err = towcard.SetBet(u.UserID, req.Key, mdr)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(mdr.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set banker failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	msg := &pbtow.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
	}
	bro := &pbtow.SetBetBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	topic.Publish(fcs.broker, bro, TopicTowCardSetBet)
	return nil
}

func (fcs *TowCardSrv) SubmitCard(ctx context.Context, req *pbtow.SubmitCardRequest,
	rsp *pbtow.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbtow.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err!=nil{
		return err
	}
	var game *mdgame.Towcard
	f := func() error {
		game, err = towcard.SubmitCard(u.UserID, mdr)
		if err != nil {
			return err
		}
		return nil
	}
	//fmt.Printf("AAASubmitCard:%+v\n",mdr)
	lock := RoomLockKey(mdr.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set banker failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	msg := &pbtow.SubmitCard{
		UserID:     u.UserID,
		GameStatus: game.Status,
		GameID:     game.GameID,
	}
	bro := &pbtow.SubmitCardBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	topic.Publish(fcs.broker, bro, TopicTowCardGameSubmitCard)
	return nil
}

func (fcs *TowCardSrv) GameResultList(ctx context.Context, req *pbtow.GameResultListRequest,
	rsp *pbtow.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := towcard.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (fcs *TowCardSrv) TowCardRecovery(ctx context.Context, req *pbtow.RecoveryRequest,
	rsp *pbtow.RecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	result, err := towcard.GameExist(req.UserID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *result
	return nil
}
