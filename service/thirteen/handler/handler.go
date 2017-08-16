package handler

import (
	"context"
	"playcards/model/room/enum"
	"playcards/model/thirteen"
	pbt "playcards/proto/thirteen"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
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

func (z *ThirteenSrv) update(gt *gsync.GlobalTimer) {
	lock := "bcr.thirteen.update.lock"
	f := func() error {
		log.Debug("thirteen update loop... and has %d thirteens")
		//now := time.Now()
		thirteen.CreateThirteen()
		return nil
	}
	gt.Register(lock, time.Second*enum.LoopTime, f)
}

func (z *ThirteenSrv) Submit(ctx context.Context, req *pbt.SubmitCards,
	rsp *pbt.SubmitCards) error {
	return nil
}

func (z *ThirteenSrv) SurrenderVote(ctx context.Context, req *pbt.Surrender,
	rsp *pbt.Surrender) error {
	return nil
}
