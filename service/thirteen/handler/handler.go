package handler

import (
	"context"
	"playcards/model/room/enum"
	"playcards/model/thirteen"
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
	lock := "bcr.thirteen.update.lock"
	f := func() error {
		log.Debug("thirteen update loop... and has %d thirteens")
		//now := time.Now()
		newGames := thirteen.CreateThirteen()
		//fmt.Printf("ThirteenUpdate:%v \n", newGames)
		if newGames != nil {
			for _, game := range newGames {
				for _, groupCard := range game.Cards {
					msg := groupCard.ToProto()
					topic.Publish(ts.broker, msg, TopicThirteenGameStart)
				}
			}
		}

		results := thirteen.CleanGame()
		//fmt.Printf("ThirteenUpdate:%v \n", newGames)
		if results != nil {
			for _, game := range results {
				msg := game.ToProto()
				topic.Publish(ts.broker, msg, TopicThirteenGameResult)
			}
		}
		return nil
	}
	gt.Register(lock, time.Second*enum.LoopTime, f)
}

func (ts *ThirteenSrv) SubmitCard(ctx context.Context, req *pbt.SubmitCard,
	rsp *pbt.SubmitCard) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rid, err := thirteen.SubmitCard(u.UserID, mdt.SubmitCardFromProto(req))
	if err != nil {
		return err
	}
	//fmt.Printf("SubmitCardSrv:%v \n", rid)
	msg := &pbt.GameReady{
		RoomID: rid,
		UserID: u.UserID,
	}
	topic.Publish(ts.broker, msg, TopicThirteenGameReady)
	return nil
}

func (ts *ThirteenSrv) SurrenderVote(ctx context.Context, req *pbt.Surrender,
	rsp *pbt.Surrender) error {
	return nil
}
