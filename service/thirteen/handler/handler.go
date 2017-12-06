package handler

import (
	"context"
	"fmt"
	"playcards/model/room"
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
)

type ThirteenSrv struct {
	server server.Server
	broker broker.Broker
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *ThirteenSrv {
	b := &ThirteenSrv{
		server: s,
		broker: s.Options().Broker,
	}
	b.update(gt)
	return b
}

func (ts *ThirteenSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.thirteen.update.lock"
	f := func() error {
		//s := time.Now()
		//log.Debug("thirteen update loop... and has %d thirteens")≤
		newGames := thirteen.CreateThirteen()
		if newGames != nil {
			for _, game := range newGames {
				for _, groupCard := range game.Cards {
					msg := groupCard.ToProto()
					msg.BankerID = game.BankerID
					topic.Publish(ts.broker, msg, TopicThirteenGameStart)
				}
			}
		}
		//e := time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("ThirteenTimesCreate:%d\n", e)

		//s = time.Now()
		games := thirteen.UpdateGame()
		if games != nil {
			for _, game := range games {
				msg := game.Result.ToProto()
				msg.Ids = game.Ids
				topic.Publish(ts.broker, msg, TopicThirteenGameResult)
			}
		}
		//e = time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("ThirteenTimesUpdate:%d\n", e)

		//s = time.Now()
		err := thirteen.CleanGame()
		if err != nil {
			log.Err("clean game loop err:%v", err)
		}
		//e = time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("ThirteenTimesClean:%d\n", e)

		return nil
	}
	gt.Register(lock, time.Millisecond*enum.LoopTime, f)
}

func (ts *ThirteenSrv) SubmitCard(ctx context.Context, req *pbt.SubmitCard,
	rsp *pbt.ThirteenReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.GetRoomByUserID(u.UserID)
	var ids []int32
	f := func() error {
		ids, err = thirteen.SubmitCard(u.UserID, mdt.SubmitCardFromProto(req, u.UserID), r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
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
		Ids:    ids,
		UserID: u.UserID,
	}
	topic.Publish(ts.broker, msg, TopicThirteenGameReady)
	//if len(pwd) == 0{
	//	return
	//}

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

func (rs *ThirteenSrv) ThirteenRecovery(ctx context.Context, req *pbt.ThirteenRequest,
	rsp *pbt.ThirteenRecoveryReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	res := &pbt.ThirteenRecoveryReply{}
	recovery, err := thirteen.ThirteenRecovery(req.RoomID, u.UserID)

	if err != nil {
		return err
	}
	res = recovery.ToProto()
	res.Result = 1
	*rsp = *res
	return nil
}
