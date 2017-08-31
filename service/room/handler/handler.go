package handler

import (
	"fmt"
	"playcards/model/room"
	enum "playcards/model/room/enum"
	enumroom "playcards/model/room/enum"
	mdr "playcards/model/room/mod"
	pbr "playcards/proto/room"
	"playcards/utils/auth"
	"playcards/utils/log"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"

	"golang.org/x/net/context"
)

type RoomSrv struct {
	server server.Server
	broker broker.Broker
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *RoomSrv {
	b := &RoomSrv{
		server: s,
		broker: s.Options().Broker,
	}
	b.update(gt)
	return b
}

func (rs *RoomSrv) update(gt *gsync.GlobalTimer) {
	lock := "bcr.room.update.lock"
	f := func() error {
		log.Debug("room update loop... and has %d rooms")
		//now := time.Now()

		rooms := room.ReInit()
		for _, room := range rooms {

			RoomResults := mdr.RoomResults{
				RoomID: room.RoomID,
				List:   room.UserResults,
			}
			msg := RoomResults.ToProto()
			//fmt.Printf("UpdateRoomSrv:%v", msg)
			topic.Publish(rs.broker, msg, TopicRoomResult)
		}
		return nil
	}
	gt.Register(lock, time.Second*enum.LoopTime, f)
}

func (rs *RoomSrv) CreateRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	r, err := room.CreateRoom(req.GameType, req.MaxNumber, req.RoundNumber,
		req.GameParam, u)
	if err != nil {
		return err
	}
	//webroom.AutoSubscribe(u.UserID)
	*rsp = *r.ToProto()
	return nil
}

func (rs *RoomSrv) EnterRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ru, r, err := room.JoinRoom(req.Password, u)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	msg := ru.ToProto()
	msg.RoomID = r.RoomID
	topic.Publish(rs.broker, msg, TopicRoomJoin)
	//webroom.AutoSubscribe(u.UserID)
	return nil
}

func (rs *RoomSrv) LeaveRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, room, err := room.LeaveRoom(u)
	if err != nil {
		return err
	}
	//webroom.AutoUnSubscribe(u.UserID)
	//*rsp = *r.ToProto()
	msg := r.ToProto()
	msg.RoomID = room.RoomID
	fmt.Printf("SSSSSSS:%d", room.Status)
	if room.Status == enumroom.RoomStatusDestroy {
		msg.Destroy = 1
	}
	topic.Publish(rs.broker, msg, TopicRoomUnJoin)
	return nil
}

func (rs *RoomSrv) SetReady(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, rid, err := room.GetReadyOrUnReady(req.Password, u.UserID, enum.UserReady)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	msg := rsp
	msg.RoomID = rid
	topic.Publish(rs.broker, msg, TopicRoomReady)
	return nil
}

// func (rs *RoomSrv) OutReady(ctx context.Context, req *pbr.Room,
// 	rsp *pbr.RoomUser) error {
// 	u, err := auth.GetUser(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	r, err := room.GetReadyOrUnReady(req.Password, u.UserID, enum.UserUnready)
// 	if err != nil {
// 		return err
// 	}
// 	*rsp = *r.ToProto()
// 	msg := rsp
// 	topic.Publish(rs.broker, msg, TopicRoomUnReady)
// 	return nil
// }

func (rs *RoomSrv) Heartbeat(ctx context.Context,
	req *pbr.Room, rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = room.Heartbeat(u.UserID)
	if err != nil {
		return err
	}
	return nil
}
