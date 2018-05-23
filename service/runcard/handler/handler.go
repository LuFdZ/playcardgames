package handler

import (
	"context"
	"fmt"
	"playcards/model/runcard"
	mdgame "playcards/model/runcard/mod"
	enumgame "playcards/model/runcard/enum"
	"playcards/model/room"
	enumroom "playcards/model/room/enum"
	pbgame "playcards/proto/runcard"
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

type RunCardSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
	glua   *lua.LState
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *RunCardSrv {
	fcs := &RunCardSrv{
		server: s,
		broker: s.Options().Broker,
		glua:   gl,
	}
	fcs.update(gt, gl)
	return fcs
}

func (rcs *RunCardSrv) update(gt *gsync.GlobalTimer, gl *lua.LState) {
	lock := "playcards.runcard.update.lock"
	f := func() error {
		rcs.count ++
		newGames := runcard.CreateGame(gl)
		if newGames != nil {
			for _, game := range newGames {
				for _, userResult := range game.GameResult.List {
					msg := &pbgame.GameStart{
						GameID:            game.GameID,
						UserID:            userResult.UserID,
						RoomStatus:        enumroom.RoomStatusStarted,
						GameStatus:        game.Status,
						OpUserID:          game.OpUserID,
						FirstUserID:       game.OpUserID,
						FirstStartGetCard: game.FirstCardType,
						UserList:          game.GetUserInfoListByID(userResult.UserID),
						CountDown: &pbgame.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SubmitCardTime,
						},
						Index: game.Index,
					}
					topic.Publish(rcs.broker, msg, TopicRunCardGameStart)
				}
				mdr, _ := cacheroom.GetRoom(game.PassWord)
				ids := mdr.GetIdsNotInGame()
				ids = append(ids, mdr.GetSuspendUser()...)
				for _, uid := range ids {
					msg := &pbgame.GameStart{
						GameID:     game.GameID,
						UserID:     uid,
						RoomStatus: enumroom.RoomStatusStarted,
						UserList:   game.GetUserInfoListByID(0),
						GameStatus: game.Status,
						CountDown: &pbgame.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumgame.SubmitCardTime,
						},
						Index: game.Index,
					}
					topic.Publish(rcs.broker, msg, TopicRunCardGameStart)
				}
			}
		}
		updateGames := runcard.UpdateGame(gl)
		if updateGames != nil {
			for _, game := range updateGames {
				mdr, _ := cacheroom.GetRoom(game.PassWord)
				ids := mdr.GetIdsNotInGame()
				ids = append(ids, mdr.GetSuspendUser()...)
				if game.Status == enumgame.GameStatusInit {
					msg := &pbgame.SubmitCardBro{
						Context: game.LastSubmitCard.ToProto(),
						Ids:     ids,
					}
					topic.Publish(rcs.broker, msg, TopicRunCardGameSubmitCard)
				} else if game.Status == enumgame.GameStatusDone {
					mdr, err := cacheroom.GetRoom(game.PassWord)
					if err != nil {
						log.Err("update game get room by password fail game:%+v,err:%+v", game, err)
						continue
					}
					msg := game.ToProto()
					bro := &pbgame.GameResultBro{
						Context: msg,
						Ids:     mdr.Ids,
					}
					bro.Ids = append(bro.Ids, mdr.GetIdsNotInGame()...)
					topic.Publish(rcs.broker, bro, TopicRunCardGameResult)
				}
			}
		}
		if rcs.count == 3 {
			err := runcard.CleanGame()
			if err != nil {
				log.Err("run card clean give up game loop err:%v", err)
			}
			rcs.count = 0
		}
		return nil
	}
	gt.Register(lock, time.Millisecond*enumgame.LoopTime, f)
}

func (rcs *RunCardSrv) SubmitCard(ctx context.Context, req *pbgame.SubmitCardRequest,
	rsp *pbgame.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}
	mdr, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	var (
		game        *mdgame.Runcard
		passUserIds []int32
	)
	f := func() error {
		game, passUserIds, err = runcard.SubmitCard(u.UserID, mdr.Password, req.CardList, rcs.glua)
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
	msg := &pbgame.SubmitCard{
		CurUserID:  u.UserID,
		GameStatus: game.Status,
		NextUserID: game.LastSubmitCard.NextUserID,
		CardType:   game.LastSubmitCard.CardType,
		CardList:   game.LastSubmitCard.CardList,
		CurCardNum: int32(len(game.GetUserInfo(u.UserID).CardList)),
		PassUserList: passUserIds,
		CountDown: &pbgame.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumgame.SubmitCardTime,
		},
	}
	bro := &pbgame.SubmitCardBro{
		Context: msg,
		Ids:     mdr.Ids,
	}
	bro.Ids = append(bro.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rcs.broker, bro, TopicRunCardGameSubmitCard)
	return nil
}

func (rcs *RunCardSrv) GameResultList(ctx context.Context, req *pbgame.GameResultListRequest,
	rsp *pbgame.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := runcard.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (rcs *RunCardSrv) RunCardRecovery(ctx context.Context, req *pbgame.RecoveryRequest,
	rsp *pbgame.RecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	result, err := runcard.GameExist(req.UserID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *result
	return nil
}
