package handler

import (
	"context"
	"fmt"
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
		s := time.Now()
		log.Debug("thirteen update loop... and has %d thirteens")
		newGames := thirteen.CreateThirteen()
		if newGames != nil {
			for _, game := range newGames {
				for _, groupCard := range game.Cards {
					msg := groupCard.ToProto()
					msg.BankerID =game.BankerID
					topic.Publish(ts.broker, msg, TopicThirteenGameStart)
				}
			}
		}

		cleanGames := thirteen.CleanGame()
		if cleanGames != nil {
			for _, game := range cleanGames {
				msg := game.ToProto()
				topic.Publish(ts.broker, msg, TopicThirteenGameResult)
			}
		}

		err := thirteen.CleanGiveUpGame()
		if err != nil {
			log.Err("clean give up game loop err:%v", err)
		}
		e := time.Now().Sub(s).Nanoseconds()
		fmt.Printf("Update times :%d", e)
		return nil
	}
	gt.Register(lock, time.Second*enum.LoopTime, f)
}

func (ts *ThirteenSrv) SubmitCard(ctx context.Context, req *pbt.SubmitCard,
	rsp *pbt.ThirteenReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rid, err := thirteen.SubmitCard(u.UserID, mdt.SubmitCardFromProto(req, u.UserID))
	if err != nil {
		return err
	}
	reply := &pbt.ThirteenReply{
		Result: 1,
	}
	*rsp = *reply
	//fmt.Printf("SubmitCardSrv:%v \n", rid)
	msg := &pbt.GameReady{
		RoomID: rid,
		UserID: u.UserID,
	}
	topic.Publish(ts.broker, msg, TopicThirteenGameReady)
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
	//fmt.Printf("get thirteen recovery:%v", recovery)
	if err != nil {
		return err
	}
	res = recovery.ToProto()
	res.Result = 1
	*rsp = *res
	return nil
}
