package handler

import (
	//"fmt"
	mdpage "playcards/model/page"
	"playcards/model/room"
	enumr "playcards/model/room/enum"
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
	//"fmt"
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
		//s := time.Now()
		log.Debug("room update loop... and has %d rooms")
		rooms := room.ReInit()
		for _, room := range rooms {
			roomResults := mdr.RoomResults{
				RoundNumber: room.RoundNumber,
				RoundNow:    room.RoundNow,
				Status:      room.Status,
				Password:    room.Password,
				List:        room.UserResults,
				CreatedAt:   room.CreatedAt,
			}
			if room.Status == enumr.RoomStatusDone {
				roomResults.List = room.UserResults
			}

			msg := roomResults.ToProto()
			msg.Ids = room.Ids
			topic.Publish(rs.broker, msg, TopicRoomResult)
		}

		giveups := room.GiveUpRoomDestroy()
		for _, giveup := range giveups {
			msg := giveup.GiveupGame.ToProto()
			msg.Ids = giveup.Ids
			topic.Publish(rs.broker, msg, TopicRoomGiveup)
		}

		room.DelayRoomDestroy()
		room.DeadRoomDestroy()
		RoomUserSocket := room.RoomUserStatusCheck()
		for _, msg := range RoomUserSocket {
			topic.Publish(rs.broker, msg, TopicRoomUserConnection)
		}
		//e := time.Now().Sub(s).Nanoseconds()
		//fmt.Printf("Update times :%d\n", e)
		return nil

	}
	gt.Register(lock, time.Millisecond*enumr.LoopTime, f)
}

func (rs *RoomSrv) CreateRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	r, err := room.CreateRoom(req.RoomType, req.GameType, req.MaxNumber,
		req.RoundNumber, req.GameParam, u, "")
	if err != nil {
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msg := r.ToProto()
	msg.UserID = u.UserID
	topic.Publish(rs.broker,msg , TopicRoomCreate)
	return nil
}

func (rs *RoomSrv) Renewal(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ids,r, err := room.RenewalRoom(req.Password, u)
	if err != nil {
		return err
	}

	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msgBack := r.ToProto()
	msgBack.UserID = r.Users[0].UserID
	topic.Publish(rs.broker, msgBack, TopicRoomCreate)

	msgAll := &pbr.RenewalRoomReady{
		//RoomID:   oldid,
		Password: r.Password,
		Status:   r.Status,
		Ids:      ids,
	}

	topic.Publish(rs.broker, msgAll, TopicRoomRenewal)
	return nil
}

func (rs *RoomSrv) EnterRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ru, r, err := room.JoinRoom(req.Password, u)
	if err != nil {
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msgBack := r.ToProto()
	msgBack.UserID = u.UserID
	topic.Publish(rs.broker, msgBack, TopicRoomCreate)

	msgAll := ru.ToProto()
	msgAll.Ids = r.Ids
	topic.Publish(rs.broker, msgAll, TopicRoomJoin)
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
	msg.Ids = room.Ids
	//msg.RoomID = room.RoomID

	if room.Status == enumr.RoomStatusDestroy {
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

	r, ids, err := room.GetReady(req.Password, u.UserID)
	if err != nil {
		return err
	}
	*rsp = *r.ToProto()
	msg := rsp
	msg.Ids = ids
	//msg.RoomID = rid
	topic.Publish(rs.broker, msg, TopicRoomReady)
	return nil
}

func (rs *RoomSrv) GiveUpGame(ctx context.Context, req *pbr.GiveUpGameRequest,
	rsp *pbr.GiveUpGameResult) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ids, result, err := room.GiveUpGame(req.Password, u.UserID)
	if err != nil {
		return err
	}
	//	*rsp = *result.ToProto()
	msg := result.ToProto()
	msg.Ids = ids
	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	return nil
}

func (rs *RoomSrv) GiveUpVote(ctx context.Context, req *pbr.GiveUpVoteRequest,
	rsp *pbr.GiveUpGameResult) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	ids, result, err := room.GiveUpVote(req.Password, req.AgreeOrNot, u.UserID)
	if err != nil {
		return err
	}
	//	*rsp = *result.ToProto()
	msg := result.ToProto()
	msg.Ids = ids
	topic.Publish(rs.broker, msg, TopicRoomGiveup)

	if result.Status == enumr.RoomStatusStarted {
		for _, userstate := range result.UserStateList {
			userstate.State = enumr.UserStateWaiting
		}
	}

	return nil
}

func (rs *RoomSrv) Shock(ctx context.Context, req *pbr.RoomUser,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	err = room.Shock(u.UserID, req.UserID)
	if err != nil {
		return err
	}
	msg := &pbr.Shock{
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
	result, roomResults, err := room.CheckRoomExist(u.UserID)
	if err != nil {
		return err
	}

	if result == 1 {
		rsp.Result = 1
		msg := roomResults.ToProto()
		msg.UserID = u.UserID
		topic.Publish(rs.broker, msg, TopicRoomExist)
	} else {
		rsp.Result = 2
		msg:=&pbr.CheckRoomExistReply{
			Result:2,
		}
		topic.Publish(rs.broker, msg, TopicRoomExist)
	}

	return nil
}

func (rs *RoomSrv) VoiceChat(ctx context.Context, req *pbr.VoiceChatRequest) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.VoiceChat(u.UserID)
	if err != nil {
		return err
	}
	msg := &pbr.VoiceChat{
		RoomID:   r.RoomID,
		UserID:   u.UserID,
		FileCode: req.FileCode,
		Ids:      r.Ids,
	}
	topic.Publish(rs.broker, msg, TopicRoomVoiceChat)
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

	//errtest := room.LuaTest()
	//if errtest != nil {
	//	return errtest
	//}
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
