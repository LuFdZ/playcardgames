package handler

import (
	"context"
	"fmt"
	"playcards/model/fourcard"
	mdgame "playcards/model/fourcard/mod"
	enumgame "playcards/model/fourcard/enum"
	"playcards/model/room"
	enumroom "playcards/model/room/enum"
	pbfour "playcards/proto/fourcard"
	"playcards/utils/auth"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
	cacheroom "playcards/model/room/cache"
	"playcards/utils/topic"
	"time"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/yuin/gopher-lua"
)

type FourCardSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *FourCardSrv {
	fcs := &FourCardSrv{
		server: s,
		broker: s.Options().Broker,
	}
	fcs.update(gt, gl)
	return fcs
}

func (fcs *FourCardSrv) update(gt *gsync.GlobalTimer, gl *lua.LState) {
	lock := "playcards.fourcard.update.lock"
	f := func() error {
		fcs.count ++
		newGames := fourcard.CreateGame()
		if newGames != nil {
			for _, game := range newGames {
				for _, userResult := range game.GameResult.List {
					msg := &pbfour.GameStart{
						GameID:     game.GameID,
						UserID:     userResult.UserID,
						BankerID:   game.BankerID,
						RoomStatus: enumroom.RoomStatusStarted,
						CardList:   userResult.CardList,
						GameStatus: game.Status,
						CountDown: &pbfour.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SetBetTime,
						},
						Index: game.Index,
					}
					topic.Publish(fcs.broker, msg, TopicFourCardGameStart)
				}
				mdr, _ := cacheroom.GetRoom(game.PassWord)
				ids := mdr.GetIdsNotInGame()
				ids = append(ids, mdr.GetSuspendUser()...)
				for _,uid := range ids{
					msg := &pbfour.GameStart{
						GameID:     game.GameID,
						UserID:     uid,
						BankerID:   game.BankerID,
						RoomStatus: enumroom.RoomStatusStarted,
						CardList:   nil,
						GameStatus: game.Status,
						CountDown: &pbfour.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SetBetTime,
						},
						Index: game.Index,
					}
					topic.Publish(fcs.broker, msg, TopicFourCardGameStart)
				}
			}
		}
		updateGames := fourcard.UpdateGame(gl)
		if updateGames != nil {
			for _, game := range updateGames {
				if game.Status == enumgame.GameStatusOrdered {
					msg := game.ToProto()
					for _, ui := range game.GameResult.List {
						var uis []*pbfour.UserInfo
						for _, uiSub := range game.GameResult.List {
							mdui := &pbfour.UserInfo{
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
						topic.Publish(fcs.broker, msg, TopicFourCardGameReady)
					}
					mdr, err := cacheroom.GetRoom(game.PassWord)
					if err != nil{
						log.Err("update game get room by password fail game:%+v,err:%+v",game,err)
						continue
					}
					msgWatch := game.ToProto()
					ids := mdr.GetIdsNotInGame()
					ids = append(ids, mdr.GetSuspendUser()...)
					for _,gu := range msgWatch.List{
						gu.CardList = nil
						gu.HeadCards = nil
						gu.TailCards = nil
					}
					for _, uid := range ids {
						msgWatch.UserID = uid
						topic.Publish(fcs.broker, msgWatch, TopicFourCardGameReady)
					}
				} else if game.Status == enumgame.GameStatusDone {
					msg := game.ToProto()
					mdr, err := cacheroom.GetRoom(game.PassWord)
					if err != nil{
						log.Err("update game get room by password fail game:%+v,err:%+v",game,err)
						continue
					}
					bro := &pbfour.GameResultBro{
						Context: msg,
						Ids:     mdr.Ids,
					}
					bro.Ids = append(bro.Ids, mdr.GetIdsNotInGame()...)
					//bro.Ids = append(bro.Ids, mdr.GetSuspendUser()...)
					topic.Publish(fcs.broker, bro, TopicFourCardGameResult)
				}
			}
		}
		if fcs.count == 3 {
			err := fourcard.CleanGame()
			if err != nil {
				log.Err("four card clean give up game loop err:%v", err)
			}
			fcs.count = 0
		}

		return nil
	}
	gt.Register(lock, time.Millisecond*enumgame.LoopTime, f)
}

func (fcs *FourCardSrv) SetBet(ctx context.Context, req *pbfour.SetBetRequest,
	rsp *pbfour.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbfour.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	f := func() error {
		err = fourcard.SetBet(u.UserID, req.Key, mdr.Password)
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
	msg := &pbfour.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
	}
	bro := &pbfour.SetBetBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	bro.Ids = append(bro.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(fcs.broker, bro, TopicFourCardSetBet)
	return nil
}

func (fcs *FourCardSrv) SubmitCard(ctx context.Context, req *pbfour.SubmitCardRequest,
	rsp *pbfour.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbfour.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err!=nil{
		return err
	}
	var game *mdgame.Fourcard
	f := func() error {
		game, err = fourcard.SubmitCard(u.UserID, mdr.Password, req.Head, req.Tail)
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
	msg := &pbfour.SubmitCard{
		UserID:     u.UserID,
		GameStatus: game.Status,
		GameID:     game.GameID,
	}
	bro := &pbfour.SubmitCardBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	bro.Ids = append(bro.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(fcs.broker, bro, TopicFourCardGameSubmitCard)
	return nil
}

func (fcs *FourCardSrv) GameResultList(ctx context.Context, req *pbfour.GameResultListRequest,
	rsp *pbfour.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := fourcard.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (fcs *FourCardSrv) FourCardRecovery(ctx context.Context, req *pbfour.RecoveryRequest,
	rsp *pbfour.RecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	result, err := fourcard.GameExist(req.UserID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *result
	return nil
}
