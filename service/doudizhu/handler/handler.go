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
	"github.com/micro/go-micro/client"
)

type DoudizhuSrv struct {
	client client.Client
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
	doudizhu.InitGoLua(gl)
	n.update(gt)
	return n
}

func (ds *DoudizhuSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.ddz.update.lock"
	f := func() error {
		newGames := doudizhu.CreateDoudizhu()
		if newGames != nil {
			for _, game := range newGames {
				msg := &pbddz.GameStart{
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

		updateGames := doudizhu.UpdateGame()
		if updateGames != nil {
			for _, game := range updateGames {
				if game.Status == enumddz.GameStatusInit {
					err := beBanker(ds.broker, game.OpID, game.GameID, enumddz.DDZNoBanker)
					if err != nil {
						log.Info("UpdateBeBankerErr userID:%d,gameID:%d,err:%v", game.OpID, game.GameID, err)
						continue
					}
				} else if game.Status == enumddz.GameStatusSubmitCard {
					err := submitCard(ds.broker, game.OpID, game.GameID, []string{})
					if err != nil {
						log.Info("UpdateSubmitCard userID:%d,gameID:%d,err:%v", game.OpID, game.GameID, err)
						continue
					}
				} else if game.Status == enumddz.GameStatusDone {
					topic.Publish(ds.broker, game.ResultToProto(), TopicDDZGameResult)
				}
			}
		}

		if ds.count == 3 {
			err := doudizhu.CleanGame()
			if err != nil {
				log.Err("clean give up game loop err:%v", err)
			}
			ds.count = 0
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
	reply.GameID = req.GameID
	*rsp = *reply

	return nil
}

func beBanker(b broker.Broker, uid int32, gid int32, getBanker int32) error {
	r, err := room.GetRoomByUserID(uid)
	if err != nil {
		return err
	}
	var game *mdddz.Doudizhu
	f := func() error {
		game, err = doudizhu.GetBanker(uid, gid, getBanker, r)
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

	msgConnet := &pbddz.BeBanker{
		GameID:       game.GameID,
		BankerStatus: game.BankerStatus,
		UserID:       uid,
		BankerType:   getBanker,
	}

	msg := &pbddz.BeBankerBro{
		Content: msgConnet,
		Ids:     game.Ids,
	}

	switch game.BankerStatus {
	case enumddz.DDZBankerStatusReStart:
		topic.Publish(b, msg, TopicDDZBeBanker)
		msg := &pbddz.GameStart{
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
		msg.Content.NextID = game.OpID
		msg.Content.CountDown = &pbddz.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumddz.GetBankerCountDown,
		}
		topic.Publish(b, msg, TopicDDZBeBanker)
		break
	case enumddz.DDZBankerStatusFinish:
		msg.Content.BankerID = game.BankerID
		msg.Content.CountDown = &pbddz.CountDown{
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

	reply := &pbddz.DefaultReply{
		Result: enumddz.Success,
	}
	err = submitCard(ds.broker, u.UserID, req.GameID, req.CardList)
	if err != nil {
		return err
	}
	//if game.Status == enumddz.GameStatusDone {
	//	topic.Publish(ds.broker, game.ResultToProto(), TopicDDZGameResult)
	//}
	reply.GameID = req.GameID
	*rsp = *reply
	return nil
}

func submitCard(b broker.Broker, uid int32, gid int32, cardList []string) error {
	r, err := room.GetRoomByUserID(uid)
	if err != nil {
		return err
	}

	var game *mdddz.Doudizhu
	f := func() error {
		game, err = doudizhu.SubmitCard(uid, gid, cardList, r)
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

	sc := &pbddz.SubmitCard{
		GameID:     game.GameID,
		SubmitID:   uid,
		CardType:   game.SubmitCardNow.CardType,
		NextID:     game.SubmitCardNow.NextID,
		ScoreTimes: game.BombTimes * game.BankerTimes * game.BaseScore,
		Status:     game.Status,
		CountDown: &pbddz.CountDown{
			ServerTime: game.OpDateAt.Unix(),
			Count:      enumddz.SubmitCardCountDown,
		},
		CardList: game.SubmitCardNow.CardList,
		//CardRemain:       uci.CardRemain,
		CardRemainNumber: int32(len(game.GetUserCard(uid).CardRemain)),
	}
	msg := &pbddz.SubmitCardBro{
		Content: sc,
		Ids:     game.Ids,
	}
	topic.Publish(b, msg, TopicDDZSubmitCard)
	return nil
}

func (ds *DoudizhuSrv) GameResultList(ctx context.Context, req *pbddz.GameResultListRequest,
	rsp *pbddz.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := doudizhu.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (ds *DoudizhuSrv) DoudizhuRecovery(ctx context.Context, req *pbddz.DoudizhuRecoveryRequest,
	rsp *pbddz.DoudizhuRecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	res, err := doudizhu.DoudizhuExist(req.UserID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *res
	return nil
}
