package handler

import (
	"context"
	"fmt"
	"playcards/model/twocard"
	mdgame "playcards/model/twocard/mod"
	enumgame "playcards/model/twocard/enum"
	"playcards/model/room"
	enumroom "playcards/model/room/enum"
	pbtwo "playcards/proto/twocard"
	"playcards/utils/auth"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/yuin/gopher-lua"
)

type TwoCardSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *TwoCardSrv {
	tcs := &TwoCardSrv{
		server: s,
		broker: s.Options().Broker,
	}
	tcs.update(gt, gl)
	return tcs
}

func (tcs *TwoCardSrv) update(gt *gsync.GlobalTimer, gl *lua.LState) {
	lock := "playcards.twocard.update.lock"
	f := func() error {
		tcs.count ++
		newGames := twocard.CreateGame()
		if newGames != nil {
			for _, game := range newGames {
				for _, userResult := range game.GameResult.List {
					msg := &pbtwo.GameStart{
						GameID:     game.GameID,
						UserID:     userResult.UserID,
						BankerID:   game.BankerID,
						RoomStatus: enumroom.RoomStatusStarted,
						CardList:   userResult.CardList,
						GameStatus: game.Status,
						CountDown: &pbtwo.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SetBetTime,
						},
					}
					topic.Publish(tcs.broker, msg, TopicTwoCardGameStart)
				}
			}
		}
		updateGames := twocard.UpdateGame(gl)
		if updateGames != nil {
			for _, game := range updateGames {
				if game.Status == enumgame.GameStatusOrdered {
					msg := game.ToProto()
					for _, ui := range game.GameResult.List {
						var uis []*pbtwo.UserInfo
						for _, uiSub := range game.GameResult.List {
							mdui := &pbtwo.UserInfo{
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
						msg.CountDown = &pbtwo.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SetBetTime,
						}
						topic.Publish(tcs.broker, msg, TopicTwoCardGameReady)
					}
				} else if game.Status == enumgame.GameStatusDone {
					msg := game.ToProto()
					bro := &pbtwo.GameResultBro{
						Context: msg,
						Ids:     game.Ids,
					}
					topic.Publish(tcs.broker, bro, TopicTwoCardGameResult)
				}
			}
		}
		if tcs.count == 3 {
			err := twocard.CleanGame()
			if err != nil {
				log.Err("two card clean give up game loop err:%v", err)
			}
			tcs.count = 0
		}

		return nil
	}
	gt.Register(lock, time.Millisecond*enumgame.LoopTime, f)
}

func (tcs *TwoCardSrv) SetBet(ctx context.Context, req *pbtwo.SetBetRequest,
	rsp *pbtwo.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbtwo.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	f := func() error {
		err = twocard.SetBet(u.UserID, req.Key, mdr)
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
	msg := &pbtwo.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
	}
	bro := &pbtwo.SetBetBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	topic.Publish(tcs.broker, bro, TopicTwoCardSetBet)
	return nil
}

func (tcs *TwoCardSrv) SubmitCard(ctx context.Context, req *pbtwo.SubmitCardRequest,
	rsp *pbtwo.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbtwo.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	var game *mdgame.Twocard
	f := func() error {
		game, err = twocard.SubmitCard(u.UserID, mdr)
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
	msg := &pbtwo.SubmitCard{
		UserID:     u.UserID,
		GameStatus: game.Status,
		GameID:     game.GameID,
	}
	bro := &pbtwo.SubmitCardBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	topic.Publish(tcs.broker, bro, TopicTwoCardGameSubmitCard)
	return nil
}

func (tcs *TwoCardSrv) GameResultList(ctx context.Context, req *pbtwo.GameResultListRequest,
	rsp *pbtwo.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := twocard.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (tcs *TwoCardSrv) TwoCardRecovery(ctx context.Context, req *pbtwo.RecoveryRequest,
	rsp *pbtwo.RecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	result, err := twocard.GameExist(req.UserID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *result
	return nil
}
