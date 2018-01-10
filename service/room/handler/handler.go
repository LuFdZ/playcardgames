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
	mdu "playcards/model/user/mod"
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

func UserLockKey(uid int32) string {
	return fmt.Sprintf("playcards.roomuser.op.lock:%d", uid)
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
	count  int32
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
		rs.count ++
		rooms := room.ReInit()
		for _, room := range rooms {
			roomResults := mdr.RoomResults{
				RoundNumber: room.RoundNumber,
				RoundNow:    room.RoundNow-1,
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
			msg.OwnerID = room.Users[0].UserID
			topic.Publish(rs.broker, msg, TopicRoomResult)
		}

		if rs.count == 30 {
			room.DelayRoomDestroy()
		}

		if rs.count == 60 {
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
		}

		if rs.count == 120 {
			room.DeadRoomDestroy()
			rs.count = 0
		}
		return nil
	}
	gt.Register(lock, time.Millisecond*enumr.LoopTime, f)

}

func (rs *RoomSrv) CreateRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	//s := time.Now()
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var r *mdr.Room
	f := func() error {
		r, err = room.CreateRoom(req.RoomType, req.GameType, req.MaxNumber,
			req.RoundNumber, req.GameParam, u, "")
		if err != nil {
			return err
		}
		return nil
	}
	lock := UserLockKey(u.UserID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s CreateRoom failed: %v", lock, err)
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}
	msg := r.ToProto()
	msg.UserID = u.UserID
	if req.RoomType != enumr.RoomTypeAgent {
		msg.OwnerID = u.UserID
	}

	msg.CreateOrEnter = enumr.CreateRoom
	topic.Publish(rs.broker, msg, TopicRoomCreate)
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
	if len(mr.Users) > 0 {
		msg.OwnerID = mr.Users[0].UserID
	} else {
		msg.OwnerID = u.UserID
	}
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
	lock := RoomLockKey(req.Password)
	f := func() error {
		oid, ids, r, err = room.RenewalRoom(req.Password, u)
		if err != nil {
			//fmt.Printf("Renewal TopicRoomRenewal:%v\n",err)
			return err
		}
		return nil
	}
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
	msgBack.OwnerID = r.Users[0].UserID
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
	if len(mr.Users) > 0 {
		msgBack.OwnerID = mr.Users[0].UserID
	} else {
		msgBack.OwnerID = u.UserID
	}
	topic.Publish(rs.broker, msgBack, TopicRoomCreate)
	msgAll := ru.ToProto()
	msgAll.OwnerID = msgBack.OwnerID
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
	r, err = cacher.GetRoomUserID(u.UserID)
	if err != nil {
		return err
	}
	if r == nil {
		return errorr.ErrUserNotInRoom
	}
	f := func() error {
		ru, r, err = room.LeaveRoom(u, r)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(r.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s leave room failed: %v", lock, err)
		return err
	}

	msg := ru.SimplyToProto()
	msg.Ids = r.Ids
	if r.Status == enumr.RoomStatusDestroy {
		msg.Destroy = 1
	}
	if len(r.Users) > 0 {
		msg.OwnerID = r.Users[0].UserID
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
			msgLeave.Diamond = mClub.Diamond //+ r.Cost
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
	msg.Ids = ids
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
	msg := result.ToProto()
	msg.Ids = ids
	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	clubDiamondTopic(rs.broker, u, mr)
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
	msg := result.ToProto()
	msg.Ids = ids
	topic.Publish(rs.broker, msg, TopicRoomGiveup)

	if result.Status == enumr.RoomStatusStarted {
		for _, userstate := range result.UserStateList {
			userstate.State = enumr.UserStateWaiting
		}
	}
	clubDiamondTopic(rs.broker, u, mr)
	return nil
}

func clubDiamondTopic(brok broker.Broker, user *mdu.User, mr *mdr.Room) {
	if mr.Status == enumr.RoomStatusGiveUp && mr.ClubID > 0 {
		if mr.ClubID > 0 {
			msg := &pbr.Room{
				RoomID: mr.RoomID,
				ClubID: mr.ClubID,
			}
			msg.UserID = 0
			mClub, _ := club.GetClubInfo(user)
			msg.Diamond = mClub.Diamond
			topic.Publish(brok, msg, TopicClubRoomFinish)
		}
	}
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

//func (rs *RoomSrv) CheckRoomExist(ctx context.Context, req *pbr.Room,
//	rsp *pbr.CheckRoomExistReply) error {
//	u, err := auth.GetUser(ctx)
//	if err != nil {
//		return err
//	}
//	result, roomResults, err := room.CheckRoomExist(u.UserID)
//	if err != nil {
//		return err
//	}
//
//	if result == 1 {
//		rsp.Result = 1
//		msg := roomResults.ToProto()
//		msg.UserID = u.UserID
//		topic.Publish(rs.broker, msg, TopicRoomExist)
//	} else {
//		rsp.Result = result
//		msg := &pbr.CheckRoomExistReply{
//			Result: result,
//		}
//		topic.Publish(rs.broker, msg, TopicRoomExist)
//	}
//
//	return nil
//}

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
		Code: errorr.ErrRoomDisband,
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

	return nil
}

func (rs *RoomSrv) GetRoomRecovery(ctx context.Context, req *pbr.Room, rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rsp.Result = 2
	hasRoom, mdroom, err := room.CheckHasRoom(u.UserID)
	if err != nil {
		return err
	}
	if hasRoom || req.RoomID > 0 {
		msg := &pbr.RoomExist{}
		msg.UserID = u.UserID
		var gameType int32
		if hasRoom {
			msg.RoomID = mdroom.RoomID
			gameType = mdroom.GameType
		} else {
			msg.RoomID = req.RoomID
			gameType = req.GameType
		}
		var top string
		switch gameType {
		case enumr.ThirteenGameType:
			top = TopicRoomThirteenExist
			break
		case enumr.NiuniuGameType:
			top = TopicRoomNiuniuExist
			break
		case enumr.DoudizhuGameType:
			top = TopicRoomDoudizhuExist
			break

		}
		topic.Publish(rs.broker, msg, top)
		rsp.Result = 1
	}
	return nil
}

func (rs *RoomSrv) GameStart(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mr, err := cacher.GetRoom(req.Password)
	if err != nil {
		return err
	}
	if mr.GameType != enumr.NiuniuGameType {
		return errorr.ErrGameType
	}
	lock := RoomLockKey(req.Password)
	f := func() error {
		err = room.GameStart(mr, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s game start failed: %v", lock, err)
		return err
	}
	msg := &pbr.Room{
		Password: mr.Password,
		ClubID:   mr.ClubID,
	}
	topic.Publish(rs.broker, msg, TopicRoomGameStart)
	return nil
}

//func (rs *RoomSrv) TestClean(ctx context.Context, req *pbr.Room, rsp *pbr.Room) error {
//	err := room.TestClean()
//	if err != nil {
//		return err
//	}
//	return nil
//}
