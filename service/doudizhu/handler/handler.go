package handler

import (
"context"
"fmt"
"time"
"playcards/model/doudizhu"
mdddz "playcards/model/doudizhu/mod"
enumddz "playcards/model/doudizhu/enum"
pbddz "playcards/proto/doudizhu"
gsync "playcards/utils/sync"
"playcards/model/room"
"playcards/utils/auth"
"playcards/utils/log"
"github.com/micro/go-micro/server"
"github.com/micro/go-micro/broker"
"github.com/yuin/gopher-lua"
"playcards/utils/topic"
)

type DoudizhuSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *DoudizhuSrv {
	n := &DoudizhuSrv{
		server: s,
		broker: s.Options().Broker,
	}
	Init()
	doudizhu.InitGoLua(gl)
	n.update(gt)
	Init()
	return n
}

func Init(brk broker.Broker) error {
	brok = brk
	if err := SubscribeAlldoudizhuMessage(brk); err != nil {
		return err
	}
	return nil
}

func SubscribeAlldoudizhuMessage(brk broker.Broker) error {
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZGameStart),
		DoudizhuGameStartHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZBeBanker),
		DoudizhuBeBankerHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZSubmitCard),
		DoudizhuSubmitCardHandler,
	)
	subscribe.SrvSubscribe(brk, topic.Topic(srvddz.TopicDDZGameResult),
		DoudizhuGameResultHandler,
	)
	return nil
}

func (ds *DoudizhuSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.ddz.update.lock"
	f := func() error {
		newGames := doudizhu.CreateDoudizhu()
		if newGames != nil {
			for _, game := range newGames {
				msg := &pbddz.DDZGameStart{
					GameID:      game.GameID,
					GetBankerID: game.OpID,
					Status:      game.Status,
					CountDown: &pbddz.CountDown{
						ServerTime: game.OpDateAt.Unix(),
						Count:      enumddz.GetBankerCountDown,
					},
				}
				for _, UserResult := range game.UserCardInfoList {
					msg.UserID = UserResult.UserID
					msg.CardList = UserResult.CardList
					topic.Publish(ds.broker, msg, TopicDDZGameStart)
				}
			}
		}

		return nil
	}

	gt.Register(lock, time.Millisecond*enumddz.LoopTime, f)
}

func (ds *DoudizhuSrv) GetBanker(ctx context.Context, req *pbddz.GetBankerRequest,
	rsp *pbddz.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbddz.DefaultReply{
		Result: enumddz.Success,
	}
	err = beBanker(ds.broker, u.UserID, req.GameID, req.GetBanker)
	if err != nil {
		return err
	}
	*rsp = *reply

	return nil
}

func beBanker(b broker.Broker, uid int32, gid int32, getBanker int32) error {
	r, err := room.GetRoomByUserID(uid)
	if err != nil {
		return err
	}

	var bankerStatus int32
	var game *mdddz.Doudizhu
	f := func() error {
		bankerStatus, game, err = doudizhu.GetBanker(uid, gid, getBanker, r)
		if err != nil {
			return err
		}
		return nil
	}

	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("doudizhu %s get banker failed: %v", lock, err)
		return err
	}

	msg := &pbddz.BeBanker{
		GameID:       game.GameID,
		BankerStatus: bankerStatus,
		UserID:       uid,
		BankerType:   getBanker,
		Ids:          game.Ids,
	}

	switch bankerStatus {
	case enumddz.DDZBankerStatusReStart:
		topic.Publish(b, msg, TopicDDZBeBanker)
		msg := &pbddz.DDZGameStart{
			GameID:      game.GameID,
			GetBankerID: game.OpID,
			Status:      game.Status,
			CountDown: &pbddz.CountDown{
				ServerTime: game.OpDateAt.Unix(),
				Count:      enumddz.GetBankerCountDown,
			},
		}
		for _, UserResult := range game.UserCardInfoList {
			msg.UserID = UserResult.UserID
			msg.CardList = UserResult.CardList
			topic.Publish(b, msg, TopicDDZGameStart)
		}
		break
	case enumddz.DDZBankerStatusContinue:
		msg.NextID = game.OpID
		msg.CountDown = &pbddz.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumddz.GetBankerCountDown,
		}
		topic.Publish(b, msg, TopicDDZBeBanker)
		break
	case enumddz.DDZBankerStatusFinish:
		msg.BankerID = game.BankerID
		msg.CountDown = &pbddz.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumddz.SubmitCardCountDown,
		}
		topic.Publish(b, msg, TopicDDZBeBanker)
		break
	}
	return nil
}

func (ds *DoudizhuSrv) SubmitCard(ctx context.Context, req *pbddz.SubmitCardRequest,
	rsp *pbddz.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.GetRoomByUserID(u.UserID)
	if err != nil {
		return err
	}
	reply := &pbddz.DefaultReply{
		Result: enumddz.Success,
	}
	var (
		remainCard []string
		submitCard *pbddz.SubmitCard
		game *mdddz.Doudizhu
	)
	f := func() error {
		remainCard,submitCard, game, err = doudizhu.SubmitCard(u.UserID, req.GameID, req.CardList, r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s submit card failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	for _,uid :=range game.Ids {
		if uid == u.UserID{
			submitCard.CardRemain = remainCard
		}else{
			submitCard.CardRemain = nil
		}
		topic.Publish(ds.broker, submitCard, TopicDDZSubmitCard)
	}

	if game.Status == enumddz.GameStatusDone{
		topic.Publish(ds.broker, game.ResultToProto(), TopicDDZGameResult)
	}
	return nil
}
