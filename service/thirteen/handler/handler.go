package handler

import (
	"context"
	"fmt"
	cacheroom "playcards/model/room/cache"
	"playcards/model/thirteen"
	"playcards/model/thirteen/enum"
	mdt "playcards/model/thirteen/mod"
	pbt "playcards/proto/thirteen"
	"playcards/utils/auth"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/yuin/gopher-lua"
)

type ThirteenSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *ThirteenSrv {
	t := &ThirteenSrv{
		server: s,
		broker: s.Options().Broker,
	}
	t.update(gt,gl)
	return t
}

func (ts *ThirteenSrv) update(gt *gsync.GlobalTimer,gl *lua.LState) {
	lock := "playcards.thirteen.update.lock"
	f := func() error {
		ts.count ++
		newGames := thirteen.CreateThirteen(gl)
		if newGames != nil {
			for _, game := range newGames {
				for _, groupCard := range game.Cards {
					msg := groupCard.ToProto()
					msg.BankerID = game.BankerID
					topic.Publish(ts.broker, msg, TopicThirteenGameStart)
				}
			}
		}
		games := thirteen.UpdateGame(gl)
		if games != nil {
			for _, game := range games {
				msg := game.Result.ToProto()
				msg.Ids = game.Ids
				topic.Publish(ts.broker, msg, TopicThirteenGameResult)
			}
		}
		if ts.count == 30 {
			err := thirteen.CleanGame()
			if err != nil {
				log.Err("clean game loop err:%v", err)
			}
			ts.count = 0
		}
		return nil
	}
	gt.Register(lock, time.Millisecond*enum.LoopTime, f)
}

func (ts *ThirteenSrv) SubmitCard(ctx context.Context, req *pbt.SubmitCard,
	rsp *pbt.ThirteenReply) error {
	//s := time.Now()
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdr, err := cacheroom.GetRoomUserID(u.UserID)
	if err != nil{
		return err
	}
	f := func() error {
		err = thirteen.SubmitCard(u.UserID, mdt.SubmitCardFromProto(req, u.UserID))
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(mdr.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return err
	}
	reply := &pbt.ThirteenReply{
		Result: 1,
	}
	*rsp = *reply

	msg := &pbt.GameReady{
		Ids:    mdr.Ids,
		UserID: u.UserID,
	}
	topic.Publish(ts.broker, msg, TopicThirteenGameReady)
	//if len(pwd) == 0{
	//	return
	//}
	//e := time.Now().Sub(s).Nanoseconds()/1000000
	//log.Info("ThirteenOpSubmitCard:%d\n", e)
	return nil
}

func (ts *ThirteenSrv) GameResultList(ctx context.Context, req *pbt.GameResultListRequest,
	rsp *pbt.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := thirteen.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (rs *ThirteenSrv) ThirteenRecovery(ctx context.Context, req *pbt.ThirteenRecoveryRequest,
	rsp *pbt.ThirteenRecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	recovery, err := thirteen.ThirteenExist(req.UserID,req.RoomID)

	if err != nil {
		return err
	}
	*rsp = *recovery
	return nil
}
