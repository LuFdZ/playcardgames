package handler

import (
	"fmt"
	"playcards/model/club"
	mdpage "playcards/model/page"
	"playcards/model/room"
	cacher "playcards/model/room/cache"
	enumr "playcards/model/room/enum"
	errorr "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	pbr "playcards/proto/room"
	//srvbill "playcards/service/bill/handler"
	//"playcards/model/bill"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilpb "playcards/utils/proto"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"

	"strings"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

func ClubRoomLockKey(clubid int32) string {
	return fmt.Sprintf("playcards.club.op.lock:%d", clubid)
}

func ClubJoinRoomLockKey(clubid int32) string {
	return fmt.Sprintf("playcards.club.op.lock:%d", clubid)
}

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
		//log.Debug("room update loop... and has %d rooms")
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
			if room.Status == enumr.RoomStatusDelay {
				if room.ClubID > 0 {
					msg := &pbr.Room{
						RoomID: room.RoomID,
						ClubID: room.ClubID,
					}
					topic.Publish(rs.broker, msg, TopicClubRoomFinish)
				}
			}
			msg := roomResults.ToProto()
			msg.Ids = room.Ids
			topic.Publish(rs.broker, msg, TopicRoomResult)
		}
		//e := time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("RoomTimesReInitUpdate:%d\n", e)

		giveups := room.GiveUpRoomDestroy()
		for _, giveup := range giveups {
			msg := giveup.GiveupGame.ToProto()
			msg.Ids = giveup.Ids
			topic.Publish(rs.broker, msg, TopicRoomGiveup)
			if giveup.Status == enumr.RoomStatusGiveUp {
				if giveup.ClubID > 0 {
					msg := &pbr.Room{
						RoomID: giveup.RoomID,
						ClubID: giveup.ClubID,
					}
					topic.Publish(rs.broker, msg, TopicClubRoomFinish)
				}
			}
		}

		//s = time.Now()
		room.DelayRoomDestroy()
		//e = time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("RoomTimesGiveUpRoomDestroy:%d\n", e)

		room.DeadRoomDestroy()

		//s = time.Now()
		RoomUserSocket := room.RoomUserStatusCheck()
		for _, msg := range RoomUserSocket {
			topic.Publish(rs.broker, msg, TopicRoomUserConnection)
		}
		//e = time.Now().Sub(s).Nanoseconds()/100000
		//log.Info("RoomTimesUserStatusCheck:%d\n", e)

		roomCount,_:= room.GetLiveRoomCount()
		log.Info("RoomTimesRoomNumber:%d\n", roomCount)
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
	msg.CreateOrEnter = enumr.CreateRoom
	topic.Publish(rs.broker, msg, TopicRoomCreate)
	//if r.RoomType == enumr.RoomTypeAgent {
	//	balance, err := bill.GetUserBalance(u.UserID)
	//	if err != nil {
	//		return err
	//	}
	//	topic.Publish(rs.broker, balance.ToProto(), srvbill.TopicBillChange)
	//}

	return nil
}

func (rs *RoomSrv) CreateClubRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.ClubID == 0 && u.ClubID != req.ClubID {
		return errorr.ErrNotClubMember
	}
	var mr *mdr.Room
	f := func() error {
		mr, err = room.CreateRoom(req.RoomType, req.GameType, req.MaxNumber,
			req.RoundNumber, req.GameParam, u, "")
		if err != nil {
			return err
		}
		return nil
	}

	lock := ClubRoomLockKey(u.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s create club room failed: %v", lock, err)
		return err
	}

	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msg := mr.ToProto()
	msg.UserID = u.UserID
	msg.CreateOrEnter = enumr.CreateRoom
	topic.Publish(rs.broker, msg, TopicRoomCreate)
	if mr.ClubID > 0 {
		mClub, _ := club.GetClubInfo(u)
		msg.Diamond = mClub.Diamond
		topic.Publish(rs.broker, msg, TopicClubRoomCreate)
	}
	return nil
}

func (rs *RoomSrv) Renewal(ctx context.Context, req *pbr.RenewalRequest,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		oid int32
		ids []int32
		r   *mdr.Room
	)
	f := func() error {
		oid, ids, r, err = room.RenewalRoom(req.Password, u)
		if err != nil {
			//fmt.Printf("Renewal TopicRoomRenewal:%v\n",err)
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s renewal failed: %v", lock, err)
		return err
	}

	msgAll := &pbr.RenewalRoomReady{
		RoomID:   oid,
		Password: r.Password,
		Status:   r.Status,
		Ids:      ids,
	}
	topic.Publish(rs.broker, msgAll, TopicRoomRenewal)
	*rsp = pbr.RoomReply{
		Result: 1,
	}

	msgBack := r.ToProto()
	msgBack.UserID = r.Users[0].UserID
	msgBack.CreateOrEnter = enumr.CreateRoom
	topic.Publish(rs.broker, msgBack, TopicRoomCreate)
	return nil
}

func (rs *RoomSrv) EnterRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ru *mdr.RoomUser
		r  *mdr.Room
	)
	mr, err := cacher.GetRoom(req.Password)
	if err != nil {
		return err
	}
	f := func() error {
		ru, r, err = room.JoinRoom(req.Password, u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	if mr.RoomType == enumr.RoomTypeClub {
		lock = ClubJoinRoomLockKey(mr.ClubID)
	}
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msgBack := r.ToProto()
	msgBack.UserID = u.UserID
	msgBack.CreateOrEnter = enumr.EnterRoom
	topic.Publish(rs.broker, msgBack, TopicRoomCreate)
	msgAll := ru.ToProto()
	//var ids []int32
	//for _, uid := range r.Ids {
	//	if uid != u.UserID {
	//		ids = append(ids, uid)
	//	}
	//}
	msgAll.Ids = r.Ids
	topic.Publish(rs.broker, msgAll, TopicRoomJoin)
	if r.ClubID > 0 {
		msgEnter := &pbr.ClubRoomUser{
			RoomID:   r.RoomID,
			UserID:   u.UserID,
			ClubID:   r.ClubID,
			Nickname: u.Nickname,
			Icon:     u.Icon,
		}
		topic.Publish(rs.broker, msgEnter, TopicClubRoomJoin)
	}
	//if r.Status == enumr.RoomStatusDiamondNoEnough {
	//	msgErr := &pbr.RoomNotice{
	//		Code: errorr.ErrDiamondNotEnough,
	//		Ids:  r.Ids,
	//	}
	//	topic.Publish(rs.broker, msgErr, TopicRoomNotice)
	//}

	return nil
}

func (rs *RoomSrv) LeaveRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ru *mdr.RoomUser
		r  *mdr.Room
	)
	f := func() error {
		ru, r, err = room.LeaveRoom(u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s leave room failed: %v", lock, err)
		return err
	}

	//*rsp = *r.ToProto()
	msg := ru.ToProto()
	msg.Ids = r.Ids
	//msg.RoomID = room.RoomID
	if r.Status == enumr.RoomStatusDestroy {
		msg.Destroy = 1
	}
	topic.Publish(rs.broker, msg, TopicRoomUnJoin)
	if r.ClubID > 0 {
		msgLeave := &pbr.Room{
			RoomID: r.RoomID,
			UserID: u.UserID,
			ClubID: r.ClubID,
		}
		topic.Publish(rs.broker, msgLeave, TopicClubRoomUnJoin)
		if r.Status == enumr.RoomStatusDestroy {
			msgLeave.UserID = 0
			mClub, _ := club.GetClubInfo(u)
			msgLeave.Diamond = mClub.Diamond + (-r.Cost)
			topic.Publish(rs.broker, msgLeave, TopicClubRoomFinish)
		}
	}
	return nil

}

func (rs *RoomSrv) SetReady(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ids []int32
		ru  *mdr.RoomUser
	)
	f := func() error {
		ru, ids, err = room.GetReady(req.Password, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set ready room failed: %v", lock, err)
		return err
	}
	*rsp = *ru.ToProto()
	msg := rsp
	//var ids []int32
	//for _, uid := range Ids {
	//	if uid != u.UserID {
	//		ids = append(ids, uid)
	//	}
	//}
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
	var (
		ids    []int32
		result *mdr.GiveUpGameResult
		mr     *mdr.Room
	)
	f := func() error {
		ids, result, mr, err = room.GiveUpGame(req.Password, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s give up game failed: %v", lock, err)
		return err
	}
	//	*rsp = *result.ToProto()
	msg := result.ToProto()
	msg.Ids = ids
	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	if mr.Status == enumr.RoomStatusGiveUp {
		if mr.ClubID > 0 {
			msg := &pbr.Room{
				RoomID: mr.RoomID,
				ClubID: mr.ClubID,
			}
			topic.Publish(rs.broker, msg, TopicClubRoomFinish)
		}
	}
	return nil
}

func (rs *RoomSrv) GiveUpVote(ctx context.Context, req *pbr.GiveUpVoteRequest,
	rsp *pbr.GiveUpGameResult) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ids    []int32
		result *mdr.GiveUpGameResult
		mr     *mdr.Room
	)
	f := func() error {
		ids, result, mr, err = room.GiveUpVote(req.Password, req.AgreeOrNot, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s give up vote room failed: %v", lock, err)
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
	if mr.Status == enumr.RoomStatusGiveUp {
		if mr.ClubID > 0 {
			msg := &pbr.Room{
				RoomID: mr.RoomID,
				ClubID: mr.ClubID,
			}
			topic.Publish(rs.broker, msg, TopicClubRoomFinish)
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
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	forwarded := ctx.Value("X-Forwarded-For")
	address := ""
	if forwarded != nil {
		addresslist, _ := forwarded.(string) //ctx.Value("X-Real-Ip").(string)
		list := strings.Split(addresslist, ",")
		address = list[0]
	}
	fb, err := room.CreateFeedback(mdr.FeedbackFromProto(req), u.UserID, address)
	if err != nil {
		return err
	}
	*rsp = *fb.ToProto()
	return nil
}

func (rs *RoomSrv) RoomResultList(ctx context.Context, req *pbr.PageRoomResultListRequest,
	rsp *pbr.RoomResultListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	page := mdpage.PageOptionFromProto(req.Page)
	roomresult, err := room.RoomResultList(page, u.UserID, req.GameType)
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
		rsp.Result = result
		msg := &pbr.CheckRoomExistReply{
			Result: result,
		}
		topic.Publish(rs.broker, msg, TopicRoomExist)
	}

	return nil
}

func (rs *RoomSrv) VoiceChat(ctx context.Context, req *pbr.VoiceChatRequest, rsp *pbr.RoomReply) error {
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

func (rs *RoomSrv) GetAgentRoomList(ctx context.Context, req *pbr.GetAgentRoomListRequest,
	rsp *pbr.GetAgentRoomListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	roomresult, err := room.GetAgentRoomList(u.UserID, req.GameType, req.Page)
	if err != nil {
		return err
	}
	*rsp = *roomresult
	return nil
}

func (rs *RoomSrv) DeleteAgentRoomRecord(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = room.DeleteAgentRoomRecord(u.UserID, req.GameType, req.RoomID, req.Password)
	if err != nil {
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}
	return nil
}

func (rs *RoomSrv) DisbandAgentRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		r *mdr.Room
	)
	f := func() error {
		r, err = room.DisbandAgentRoom(u.UserID, req.Password)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s disband agent room failed: %v", lock, err)
		return err
	}
	msg := &pbr.RoomNotice{
		Code: errorr.ErrDiamondNotEnough,
		Ids:  r.Ids,
	}

	topic.Publish(rs.broker, msg, TopicRoomNotice)

	return nil
}

func (rs *RoomSrv) GetRoomUserLocation(ctx context.Context, req *pbr.Room, rsp *pbr.RoomUsersReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rus, err := room.GetRoomUserLocation(u)
	if err != nil {
		return err
	}
	rsp.List = rus
	return nil
}

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
