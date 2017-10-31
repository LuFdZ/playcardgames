package handler

import (
	"context"
	"playcards/model/niuniu"
	cacheniu "playcards/model/niuniu/cache"
	enumniu "playcards/model/niuniu/enum"
	enumr "playcards/model/room/enum"
	pbniu "playcards/proto/niuniu"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilproto "playcards/utils/proto"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
)

type NiuniuSrv struct {
	server server.Server
	broker broker.Broker
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *NiuniuSrv {
	n := &NiuniuSrv{
		server: s,
		broker: s.Options().Broker,
	}
	n.Update(gt)
	return n
}

func (ns *NiuniuSrv) Update(gt *gsync.GlobalTimer) {
	lock := "playcards.niu.update.lock"

	f := func() error {
		//s := time.Now()
		log.Debug("niuniu update loop... and has %d niunius")

		newGames := niuniu.CreateNiuniu()
		if newGames != nil {
			for _, game := range newGames {
				for _, UserResult := range game.Result.List {
					//fmt.Printf("Update Game New Game:%v", game)
					cardlist := UserResult.Cards.
						CardList[:len(UserResult.Cards.CardList)-1]
					msg := &pbniu.NiuniuGameStart{
						Role:       0, //UserResult.Info.Role
						UserID:     UserResult.UserID,
						BankerID:   game.BankerID,
						RoomStatus: enumr.RoomStatusStarted,
						CardList:   cardlist,
						GameStatus: game.Status,
					}
					topic.Publish(ns.broker, msg, TopicNiuniuGameStart)
				}
			}
		}
		//sub := time.Now().Sub(*niuniu.OpDateAt)
		updateGames := niuniu.UpdateGame()
		if updateGames != nil {
			for _, game := range updateGames {
				if game.BroStatus == enumniu.GameStatusCountDown {
					//fmt.Printf("1111 Game Status Init time")
					refresh := int32(time.Now().Sub(*game.RefreshDateAt).Seconds())
					if refresh < 1 {
						continue
					}
					sub := int32(time.Now().Sub(*game.OpDateAt).Seconds())
					if sub > 1 {
						var totalTime int32
						if game.Status < enumniu.GameStatusGetBanker {
							totalTime = enumniu.GetBankerTime
						} else if game.Status < enumniu.GameStatusAllSetBet {
							totalTime = enumniu.SetBetTime
						} else if game.Status < enumniu.GameStatusStarted {
							totalTime = enumniu.SubmitCardTime
						}
						countDown := totalTime - sub
						//fmt.Printf("2222 Game Status Count Down:%d", countDown)
						if countDown > 0 {
							msg := &pbniu.CountDown{
								//RoomID: game.RoomID,
								//Status: enumniu.ToBetScoreMap[game.Status],
								Ids:game.Ids,
								Time: int32(countDown),
							}
							topic.Publish(ns.broker, msg, TopicNiuniuCountDown)
							now := gorm.NowFunc()
							game.SubDateAt = &now
							game.RefreshDateAt = &now
							err := cacheniu.UpdateGame(game)
							if err != nil {
								log.Err("niuniu set session failed, %v", err)
								return nil
							}
						}
					}
				} else if game.BroStatus == enumniu.GameStatusGetBanker {
					if game.HasNewBanker {
						msg := &pbniu.BeBanker{
							BankerID:   game.BankerID,
							GameStatus: enumniu.ToBetScoreMap[game.Status],
							Ids:        game.Ids,
							//RoomID:     game.RoomID,
						}
						utilproto.ProtoSlice(game.GetBankerList, &msg.List)
						topic.Publish(ns.broker, msg, TopicNiuniuBeBanker)
					}
				} else if game.BroStatus == enumniu.GameStatusAllSetBet {
					for _, UserResult := range game.Result.List {
						msg := &pbniu.AllBet{
							UserID: UserResult.UserID,
							Status: enumniu.ToBetScoreMap[game.Status],
							Card:   UserResult.Cards.CardList[4],
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
		err := niuniu.CleanGame()
		if err != nil {
			log.Err("clean give up game loop err:%v", err)
		}
		//e := time.Now().Sub(s).Nanoseconds()
		//fmt.Printf("Update times :%d", e)
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
	_, err = niuniu.GetBanker(u.UserID, req.Key)
	if err != nil {
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
	ids, err := niuniu.SetBet(u.UserID, req.Key)
	if err != nil {
		return err
	}

	*rsp = *reply
	msg := &pbniu.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
		Ids:ids,
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
	ids, err := niuniu.SubmitCard(u.UserID)
	if err != nil {
		return err
	}
	*rsp = *reply
	msg := &pbniu.GameReady{
		UserID: u.UserID,
		Ids: ids,
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

func (ns *NiuniuSrv) NiuniuRecovery(ctx context.Context, req *pbniu.NiuniuRequest,
	rsp *pbniu.NiuniuReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	status, recovery, err := niuniu.NiuniuRecovery(req.RoomID, u.UserID)
	if err != nil {
		return err
	}
	res := &pbniu.NiuniuReply{
		Result: recovery.ToProto(),
	}
	res.Result.Status = enumniu.ToBetScoreMap[status]
	*rsp = *res
	return nil
}
