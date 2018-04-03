package handler

import (
	"fmt"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	gsync "playcards/utils/sync"
	enumroom "playcards/model/room/enum"
	enumgroom "playcards/model/goldroom/enum"
	pbgroom "playcards/proto/goldroom"
	pbroom "playcards/proto/room"
	errgroom "playcards/model/goldroom/errors"
	"playcards/utils/auth"
	"playcards/model/goldroom"
	"time"
	"context"
	"playcards/utils/topic"
)

func RoomLockKey(rid int32) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", rid)
}

func UserLockKey(uid int32) string {
	return fmt.Sprintf("playcards.roomuser.op.lock:%d", uid)
}

type GoldRoomSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *GoldRoomSrv {
	r := &GoldRoomSrv{
		server: s,
		broker: s.Options().Broker,
	}
	r.update(gt)
	return r
}

func (grs *GoldRoomSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.room.update.lock"
	f := func() error {
		grs.count ++
		rus := goldroom.UpdateGoldRoom()
		for msg, opIndex := range rus {
			switch opIndex {
			case enumgroom.UserOpJoin: //机器人操作值 加入游戏
				topic.Publish(grs.broker, msg, TopicRoomJoin)
				break
			case enumgroom.UserOpReady: //机器人操作值 游戏准备
				topic.Publish(grs.broker, msg, TopicRoomReady)
				break
			case enumgroom.UserOpRemove: //清除不活动玩家
				topic.Publish(grs.broker, msg, TopicRoomUnJoin)
				msg := &pbroom.RoomNotice{
					Code: errgroom.ErrPlayerRemoved,
					Ids:  msg.Ids,
				}
				topic.Publish(grs.broker, msg, TopicRoomNotice)
				break
			}
		}
		return nil
	}
	gt.Register(lock, time.Millisecond*enumgroom.LoopTime, f)
}

func (grs *GoldRoomSrv) EnterRoom(ctx context.Context, req *pbgroom.EnterRoomRequest,
	rsp *pbgroom.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdru, mdroom, err := goldroom.JoinRoom(req.GameType, req.Level, u)
	if err != nil {
		return err
	}

	*rsp = pbgroom.DefaultReply{
		Result: 1,
	}
	msgBack := mdroom.ToProto()
	msgBack.UserID = u.UserID
	//msgBack.UserRole = enumroom.
	msgBack.CreateOrEnter = enumroom.EnterRoom
	topic.Publish(grs.broker, msgBack, TopicRoomCreate)
	msgAll := mdru.ToProto()
	msgAll.Ids = mdroom.Ids
	topic.Publish(grs.broker, msgAll, TopicRoomJoin)
	return nil
}

func (grs *GoldRoomSrv) LeaveRoom(ctx context.Context, req *pbgroom.LeaveRoomRequest,
	rsp *pbgroom.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ru, mdr, err := goldroom.LeaveRoom(u)
	if err != nil {
		return err
	}
	*rsp = pbgroom.DefaultReply{
		Result: 1,
	}
	msg := ru.SimplyToProto()
	msg.Ids = mdr.Ids
	if mdr.Status == enumroom.RoomStatusDestroy {
		msg.Destroy = 1
	}
	if len(mdr.Users) > 0 {
		msg.OwnerID = mdr.Users[0].UserID
		msg.BankerID = msg.OwnerID
	}
	topic.Publish(grs.broker, msg, TopicRoomUnJoin)
	return nil

}

func (grs *GoldRoomSrv) SetReady(ctx context.Context, req *pbgroom.SetReadyRequest,
	rsp *pbgroom.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ids,err := goldroom.SetReady(u.UserID,req.Password)
	if err != nil {
		return err
	}
	*rsp = pbgroom.DefaultReply{
		Result: 1,
	}

	msg := &pbroom.RoomUser{UserID: u.UserID}
	msg.Ids = ids
	topic.Publish(grs.broker, msg, TopicRoomReady)
	return nil
}
