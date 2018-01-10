package handler

import (
	"context"
	"fmt"
	"playcards/model/niuniu"
	enumniu "playcards/model/niuniu/enum"
	"playcards/model/room"
	enumr "playcards/model/room/enum"
	pbniu "playcards/proto/niuniu"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilproto "playcards/utils/proto"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"github.com/yuin/gopher-lua"
)

type NiuniuSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer, gl *lua.LState) *NiuniuSrv {
	n := &NiuniuSrv{
		server: s,
		broker: s.Options().Broker,
	}
	//niuniu.InitGoLua(gl)
	n.update(gt, gl)
	return n
}

func (ns *NiuniuSrv) update(gt *gsync.GlobalTimer, gl *lua.LState) {
	lock := "playcards.niu.update.lock"

	f := func() error {
		ns.count ++
		//s := time.Now()
		//log.Debug("niuniu update loop... and has %d niunius")
		newGames := niuniu.CreateNiuniu(gl)
		if newGames != nil {
			for _, game := range newGames {
				for _, UserResult := range game.Result.List {
					cardlist := UserResult.Cards.
						CardList[:len(UserResult.Cards.CardList)-1]
					msg := &pbniu.NiuniuGameStart{
						Role:       0,
						UserID:     UserResult.UserID,
						BankerID:   game.BankerID,
						RoomStatus: enumr.RoomStatusStarted,
						CardList:   cardlist,//[]string{"4_11","3_11","2_1","4_4"},//
						GameStatus: game.Status,
						CountDown: &pbniu.CountDown{
							ServerTime: game.OpDateAt.Unix(),
							Count:      enumniu.GetBankerTime,
						},
					}
					topic.Publish(ns.broker, msg, TopicNiuniuGameStart)
				}
			}
		}
		//sub := time.Now().Sub(*niuniu.OpDateAt)
		updateGames := niuniu.UpdateGame(gl)
		if updateGames != nil {
			for _, game := range updateGames {
				//if game.BroStatus == enumniu.GameStatusCountDown {
				//	//fmt.Printf("1111 Game Status Init time")
				//	refresh := int32(time.Now().Sub(*game.RefreshDateAt).Seconds())
				//	if refresh < 1 {
				//		continue
				//	}
				//	sub := int32(time.Now().Sub(*game.OpDateAt).Seconds())
				//	if sub > 1 {
				//		var totalTime int32
				//		if game.Status < enumniu.GameStatusGetBanker {
				//			totalTime = enumniu.GetBankerTime
				//		} else if game.Status < enumniu.GameStatusAllSetBet {
				//			totalTime = enumniu.SetBetTime
				//		} else if game.Status < enumniu.GameStatusStarted {
				//			totalTime = enumniu.SubmitCardTime
				//		}
				//		countDown := totalTime - sub
				//		//fmt.Printf("2222 Game Status Count Down:%d", countDown)
				//		if countDown > 0 {
				//			msg := &pbniu.CountDown{
				//				//RoomID: game.RoomID,
				//				//Status: enumniu.ToBetScoreMap[game.Status],
				//				Ids:  game.Ids,
				//				Time: int32(countDown),
				//			}
				//			topic.Publish(ns.broker, msg, TopicNiuniuCountDown)
				//			now := gorm.NowFunc()
				//			game.SubDateAt = &now
				//			game.RefreshDateAt = &now
				//			err := cacheniu.UpdateGame(game)
				//			if err != nil {
				//				log.Err("niuniu set session failed, %v", err)
				//				return nil
				//			}
				//		}
				//	}
				//} else
				if game.BroStatus == enumniu.GameStatusGetBanker {
					if game.HasNewBanker {
						msg := &pbniu.BeBanker{
							BankerID:   game.BankerID,
							GameStatus: enumniu.ToBetScoreMap[game.Status],
							Ids:        game.Ids,
							CountDown:  &pbniu.CountDown{game.OpDateAt.Unix(), enumniu.SetBetTime},
						}
						utilproto.ProtoSlice(game.GetBankerList, &msg.List)
						topic.Publish(ns.broker, msg, TopicNiuniuBeBanker)
					}
				} else if game.BroStatus == enumniu.GameStatusAllSetBet {
					for _, UserResult := range game.Result.List {
						msg := &pbniu.AllBet{
							UserID:    UserResult.UserID,
							Status:    enumniu.ToBetScoreMap[game.Status],
							Card:      UserResult.Cards.CardList[4],
							CountDown: &pbniu.CountDown{game.OpDateAt.Unix(), enumniu.SubmitCardTime},
						}
						topic.Publish(ns.broker, msg, TopicNiuniuAllBet)
					}
				} else if game.BroStatus == enumniu.GameStatusDone {
					msg := game.Result.ToProto()
					msg.Status = enumniu.ToBetScoreMap[game.Status]
					msg.Ids = game.Ids
					topic.Publish(ns.broker, msg, TopicNiuniuGameResult)
					// game.Status = enumniu.GameStatusDone
					// niuniu.UpdateNiuniu(game, false)
				}
			}
		}
		if ns.count == 3 {
			err := niuniu.CleanGame()
			if err != nil {
				log.Err("clean give up game loop err:%v", err)
			}
			//e := time.Now().Sub(s).Nanoseconds()/1000000
			//fmt.Printf("niuniu Update times :%d\n", e)
			ns.count = 0
		}

		return nil
	}
	gt.Register(lock, time.Millisecond*enumniu.LoopTime, f)
}

func (ns *NiuniuSrv) GetBanker(ctx context.Context, req *pbniu.GetBankerRequest,
	rsp *pbniu.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbniu.DefaultReply{
		Result: enumniu.Success,
	}
	r, err := room.GetRoomByUserID(u.UserID)
	f := func() error {
		err = niuniu.GetBanker(u.UserID, req.Key, r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s get banker failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	return nil
}

func (ns *NiuniuSrv) SetBet(ctx context.Context, req *pbniu.SetBetRequest,
	rsp *pbniu.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbniu.DefaultReply{
		Result: enumniu.Success,
	}
	r, err := room.GetRoomByUserID(u.UserID)
	var ids []int32
	f := func() error {
		ids, err = niuniu.SetBet(u.UserID, req.Key, r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set banker failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	msg := &pbniu.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
		Ids:    ids,
	}
	topic.Publish(ns.broker, msg, TopicNiuniuSetBet)
	return nil
}

func (ns *NiuniuSrv) SubmitCard(ctx context.Context, req *pbniu.SubmitCardRequest,
	rsp *pbniu.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	reply := &pbniu.DefaultReply{
		Result: enumniu.Success,
	}
	r, err := room.GetRoomByUserID(u.UserID)
	var ids []int32
	f := func() error {
		ids, err = niuniu.SubmitCard(u.UserID, r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set banker failed: %v", lock, err)
		return err
	}
	*rsp = *reply
	msg := &pbniu.GameReady{
		UserID: u.UserID,
		Ids:    ids,
	}
	topic.Publish(ns.broker, msg, TopicNiuniuGameReady)
	return nil
}

func (ns *NiuniuSrv) GameResultList(ctx context.Context, req *pbniu.GameResultListRequest,
	rsp *pbniu.GameResultListReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	results, err := niuniu.GameResultList(req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *results
	return nil
}

func (ns *NiuniuSrv) NiuniuRecovery(ctx context.Context, req *pbniu.NiuniuRecoveryRequest,
	rsp *pbniu.NiuniuRecoveryReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	result, err := niuniu.NiuniuExist(req.UserID,req.RoomID)
	if err != nil {
		return err
	}
	*rsp = *result
	return nil
}
