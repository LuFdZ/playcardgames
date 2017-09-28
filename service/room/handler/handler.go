package handler

import (
	mdpage "playcards/model/page"
	"playcards/model/room"
	enum "playcards/model/room/enum"
	enumroom "playcards/model/room/enum"
	mdr "playcards/model/room/mod"
	pbr "playcards/proto/room"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilpb "playcards/utils/proto"
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
	r := &RoomSrv{
		server: s,
		broker: s.Options().Broker,
	}
	r.update(gt)
	return r
}

func (rs *RoomSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.room.update.lock"
	f := func() error {
		log.Debug("room update loop... and has %d rooms")
		//now := time.Now()

		rooms := room.ReInit()
		//fmt.Printf("rooms update :%d", len(rooms))
		for _, room := range rooms {
			roomResults := mdr.RoomResults{
				RoomID:      room.RoomID,
				RoundNumber: room.RoundNumber,
				RoundNow:    room.RoundNow,
				Status:      room.Status,
				Password:    room.Password,
				List:        room.UserResults,
			}
			if room.Status == enum.RoomStatusDone {
				roomResults.List = room.UserResults
			}

			msg := roomResults.ToProto()
			//fmt.Printf("UpdateRoomSrv:%v", room.UserResults)
			topic.Publish(rs.broker, msg, TopicRoomResult)
		}

		err := room.RoomDestroy()
		if err != nil {
			log.Err("room destroy loop:%v", err)
		}
		RoomUserSocket := room.RoomUserStatusCheck()
		for _, msg := range RoomUserSocket {
			topic.Publish(rs.broker, msg, TopicRoomUserConnection)
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

	r, err := room.CreateRoom(req.RoomType, req.GameType, req.MaxNumber,
		req.RoundNumber, req.GameParam, u, "")
	if err != nil {
		return err
	}
	//webroom.AutoSubscribe(u.UserID)
	*rsp = *r.ToProto()
	return nil
}

func (rs *RoomSrv) Renewal(ctx context.Context, req *pbr.Room,
	rsp *pbr.Room) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	oldid, r, err := room.RenewalRoom(req.Password, u)
	if err != nil {
		return err
	}

	//webroom.AutoSubscribe(u.UserID)
	*rsp = *r.ToProto()
	msg := &pbr.RenewalRoomReady{
		RoomID:   oldid,
		Password: r.Password,
		Status:   r.Status,
	}

	topic.Publish(rs.broker, msg, TopicRoomRenewal)
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

	r, rid, err := room.GetReady(req.Password, u.UserID)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	msg := rsp
	msg.RoomID = rid
	topic.Publish(rs.broker, msg, TopicRoomReady)
	return nil
}

func (rs *RoomSrv) GiveUpGame(ctx context.Context, req *pbr.GiveUpGameRequest,
	rsp *pbr.GiveUpGameResult) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	result, err := room.GiveUpGame(req.Password, req.AgreeOrNot, u.UserID)
	if err != nil {
		return err
	}
	//	*rsp = *result.ToProto()
	msg := result.ToProto()

	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	return nil
}

func (rs *RoomSrv) Shock(ctx context.Context, req *pbr.RoomUser,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	rid, err := room.Shock(u.UserID, req.UserID)
	if err != nil {
		return err
	}
	msg := &pbr.Shock{
		RoomID:     rid,
		UserIDFrom: u.UserID,
		UserIDTo:   req.UserID,
	}
	topic.Publish(rs.broker, msg, TopicRoomShock)
	return nil
}

func (rs *RoomSrv) PageFeedbackList(ctx context.Context,
	req *pbr.PageFeedbackListRequest, rsp *pbr.PageFeedbackListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	l, rows, err := room.PageFeedbackList(page,
		mdr.FeedbackFromProto(req.Feedback))
	if err != nil {
		return err
	}

	err = utilpb.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	return nil
}

func (rs *RoomSrv) CreateFeedback(ctx context.Context, req *pbr.Feedback,
	rsp *pbr.Feedback) error {

	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	fb, err := room.CreateFeedback(mdr.FeedbackFromProto(req))
	if err != nil {
		return err
	}
	*rsp = *fb.ToProto()

	return nil
}

func (rs *RoomSrv) RoomResultList(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomResultListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	roomresult, err := room.RoomResultList(u.UserID, req.GameType)
	if err != nil {
		return err
	}
	*rsp = *roomresult
	return nil
}

func (rs *RoomSrv) CheckRoomExist(ctx context.Context, req *pbr.Room,
	rsp *pbr.CheckRoomExistReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	//GiveupResult
	res := &pbr.CheckRoomExistReply{}
	room, err := room.CheckRoomExist(u.UserID)
	if err != nil {
		return err
	}
	res.Room = room.ToProto()
	if room.Status == enum.RoomStatusWaitGiveUp {
		res.GiveupResult = room.GiveupGame.ToProto()
	}

	roomResults := mdr.RoomResults{
		RoomID:      room.RoomID,
		RoundNumber: room.RoundNumber,
		RoundNow:    room.RoundNow,
		Status:      room.Status,
		Password:    room.Password,
		List:        room.UserResults,
	}
	if room.Status == enum.RoomStatusDone {
		roomResults.List = room.UserResults
	}
	res.GameResult = roomResults.ToProto()

	*rsp = *res

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

	errtest := room.LuaTest()
	if errtest != nil {
		return errtest
	}
	// u, err := auth.GetUser(ctx)
	// if err != nil {
	// 	return err
	// }
	// err = room.Heartbeat(u.UserID)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (rs *RoomSrv) TestClean(ctx context.Context, req *pbr.Room, rsp *pbr.Room) error {
	err := room.TestClean()
	if err != nil {
		return err
	}
	return nil
}
