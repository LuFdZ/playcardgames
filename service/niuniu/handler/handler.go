package handler

import (
	"context"
	"fmt"
	"playcards/model/niuniu"
	enumniu "playcards/model/niuniu/enum"
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
)

type NiuniuSrv struct {
	server server.Server
	broker broker.Broker
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *NiuniuSrv {
	b := &NiuniuSrv{
		server: s,
		broker: s.Options().Broker,
	}
	b.Update(gt)

	//enumniu.BankerScoreMap = *make(map[int32]int32)
	enumniu.BankerScoreMap = map[int32]int32{
		1: 0,
		2: 1,
		3: 2,
		4: 3,
		5: 4,
	}

	enumniu.BetScoreMap = map[int32]int32{
		1: 5,
		2: 10,
		3: 15,
		4: 20,
		5: 25,
	}

	return b
}

func (ns *NiuniuSrv) Update(gt *gsync.GlobalTimer) {
	lock := "playcards.niuniu.update.lock"
	f := func() error {
		log.Debug("niuniu update loop... and has %d niunius")

		newGames := niuniu.CreateNiuniu()
		if newGames != nil {
			for _, game := range newGames {
				for _, UserResult := range game.Result.List {
					//fmt.Printf("Update Game New Game:%v", game)
					cardlist := UserResult.Cards.CardList[:len(UserResult.Cards.CardList)-1]
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
				if game.Status == enumniu.GameStatusInit {
					fmt.Printf("1111 Game Status Init time")
					sub := time.Now().Sub(*game.OpDateAt)
					countDown := enumniu.GetBankerTime - sub.Seconds()
					msg := &pbniu.CountDown{
						RoomID: game.RoomID,
						Status: game.Status,
						Time:   int32(countDown),
					}
					topic.Publish(ns.broker, msg, TopicNiuniuCountDown)
				} else if game.Status == enumniu.GameStatusBeBanker {
					msg := &pbniu.BeBanker{
						BankerID:   game.BankerID,
						GameStatus: game.Status,
						RoomID:     game.RoomID,
					}
					utilproto.ProtoSlice(game.GetBankerList, &msg.List)
					topic.Publish(ns.broker, msg, TopicNiuniuBeBanker)
				} else if game.Status == enumniu.GameStatusSetBet {
					sub := time.Now().Sub(*game.OpDateAt)
					countDown := enumniu.GetBankerTime - sub.Seconds()
					msg := &pbniu.CountDown{
						RoomID: game.RoomID,
						Status: game.Status,
						Time:   int32(countDown),
					}
					topic.Publish(ns.broker, msg, TopicNiuniuCountDown)
				} else if game.Status == enumniu.GameStatusStarted {
					msg := &pbniu.AllBet{
						RoomID: game.RoomID,
						Status: game.Status,
					}
					topic.Publish(ns.broker, msg, TopicNiuniuAllBet)
				} else if game.Status == enumniu.GameStatusDone {
					msg := game.Result.ToProto()
					topic.Publish(ns.broker, msg, TopicNiuniuGameResult)
				}
			}
		}
		return nil
	}
	gt.Register(lock, time.Second*enumniu.LoopTime, f)
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
	fmt.Printf("AAAddd User Get Banker%d", req.Key)
	_, err = niuniu.GetBanker(u.UserID, req.Key)
	if err != nil {
		return err
	}
	*rsp = *reply
	// msg := &pbniu.GetBanker{
	// 	UserID: u.UserID,
	// 	Value:  req.Value,
	// 	RoomID: rid,
	// }
	// topic.Publish(ns.broker, msg, TopicNiuniuGetBanker)
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
	rid, err := niuniu.SetBet(u.UserID, req.Key)
	if err != nil {
		return err
	}

	*rsp = *reply
	msg := &pbniu.SetBet{
		UserID: u.UserID,
		Key:    req.Key,
		RoomID: rid,
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
	rid, err := niuniu.SubmitCard(u.UserID)
	if err != nil {
		return err
	}
	*rsp = *reply
	msg := &pbniu.GameReady{
		UserID: u.UserID,
		RoomID: rid,
	}
	topic.Publish(ns.broker, msg, TopicNiuniuGameReady)
	return nil
}