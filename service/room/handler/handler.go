package handler

import (
	"fmt"
	"playcards/model/club"
	mdpage "playcards/model/page"
	"playcards/model/room"
	cacher "playcards/model/room/cache"
	enumr "playcards/model/room/enum"
	errorr "playcards/model/room/errors"
	mdroom "playcards/model/room/mod"
	pbr "playcards/proto/room"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilpb "playcards/utils/proto"
	gsync "playcards/utils/sync"
	pbmail "playcards/proto/mail"
	srvmail "playcards/service/mail/handler"
	enummail "playcards/model/mail/enum"
	enumclub "playcards/model/club/enum"
	"playcards/utils/topic"
	"time"
	"strings"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
	"bcr/utils/errors"
)

// 房间操作锁
func RoomLockKey(pwd string) string {
	return fmt.Sprintf("playcards.room.op.lock:%s", pwd)
}

//用户操作锁
func UserLockKey(uid int32) string {
	return fmt.Sprintf("playcards.roomuser.op.lock:%d", uid)
}

//俱乐部操作锁
func ClubRoomLockKey(clubid int32) string {
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
	cacher.SetConfigKey()
	return r
}

func (rs *RoomSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.room.update.lock"
	f := func() error {
		rs.count ++
		room.ShuffleDelay()
		//shuffleRooms := room.ShuffleDelay()
		//for _, mdr := range shuffleRooms {
		//	if mdr.Shuffle > 0 {
		//		msg := &pbr.ShuffleCardBro{
		//			UserID: mdr.Shuffle,
		//			Ids:    mdr.Ids,
		//		}
		//		topic.Publish(rs.broker, msg, TopicRoomGShuffleCardBro)
		//}
		rooms := room.ReInit()
		for _, room := range rooms {
			roomResults := mdroom.RoomResults{
				RoundNumber: room.RoundNumber,
				RoundNow:    room.RoundNow,
				Status:      room.Status,
				Password:    room.Password,
				List:        room.UserResults,
				RoomType:    room.RoomType,
				SubRoomType: room.SubRoomType,
				CreatedAt:   room.CreatedAt,
			}
			//if room.Shuffle > 0 {
			//	msg := &pbr.ShuffleCardBro{
			//		UserID: room.Shuffle,
			//		Ids:    room.Ids,
			//	}
			//	topic.Publish(rs.broker, msg, TopicRoomGShuffleCardBro)
			//} else
			//fmt.Printf("AAAA Update ReInit:%d|%d\n",room.RoundNow,roomResults.RoundNow)

			if room.Status == enumr.RoomStatusDone {
				roomResults.List = room.UserResults
			}
			if room.Status == enumr.RoomStatusDelay {
				if room.ClubID > 0 {
					msg := &pbr.Room{
						RoomID: room.RoomID,
						ClubID: room.ClubID,
					}
					mClub, _ := club.GetClubFromDB(room.ClubID)
					msg.Diamond = mClub.Diamond //+ r.Cost
					topic.Publish(rs.broker, msg, TopicClubRoomFinish)
				}
			}
			if room.RoomType == enumr.RoomTypeClub && room.SubRoomType == enumr.SubTypeClubMatch {
				for _, ru := range roomResults.List {
					mcm, err := club.GetClubMember(room.ClubID, ru.UserID)
					if err != nil {
						log.Err("reinit get clubmember err:%v", err)
						continue
					}
					ru.ClubCoin = mcm.ClubCoin
				}
			}
			msg := roomResults.ToProto()
			for _, ru := range room.Users {
				mdru := &mdroom.RoomUser{
					UserID:   ru.UserID,
					UserRole: ru.UserRole,
				}
				msg.UserList = append(msg.UserList, mdru.SimplyToProto())
			}
			msg.Ids = room.Ids
			msg.Ids = append(msg.Ids, room.GetIdsNotInGame()...)
			msg.OwnerID = room.Users[0].UserID
			//fmt.Printf("UpdateRoomResult:%+v\n", msg)
			topic.Publish(rs.broker, msg, TopicRoomResult)

			if room.RoomNoticeCode == enumr.BankerClubCoinNoEnough {
				mdclub, _ := club.GetClubFromDB(room.ClubID)

				mderr := errors.Parse(errorr.ErrBankerClubCoinNoEnough.Error())
				if mdclub != nil {
					mderr.Detail = fmt.Sprintf(mderr.Detail, mdclub.Setting.ClubCoinBaseScore)
				}

				msg := &pbr.RoomNotice{
					Code:    int32(mderr.Code),
					Message: mderr.Detail,
					Ids:     room.Ids,
				}
				msg.Ids = append(msg.Ids, room.GetIdsNotInGame()...)
				topic.Publish(rs.broker, msg, TopicRoomNotice)
			}

			//msgNotice := &pbr.RoomNotice{
			//	Code: errorr.ErrRoomDisband,
			//	Ids:  r.Ids,
			//}
			//msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
			//topic.Publish(rs.broker, msg, TopicRoomNotice)
		}

		room.AutoReady()

		if rs.count == 30 {
			room.DelayRoomDestroy()
		}

		if rs.count == 10 {
			giveups := room.GiveUpRoomDestroy()
			for _, giveup := range giveups {
				msg := giveup.GiveupGame.ToProto()
				msg.Ids = giveup.Ids
				msg.Ids = append(msg.Ids, giveup.GetIdsNotInGame()...)
				topic.Publish(rs.broker, msg, TopicRoomGiveup)
				if giveup.Status == enumr.RoomStatusGiveUp {
					if giveup.ClubID > 0 {
						msg := &pbr.Room{
							RoomID: giveup.RoomID,
							ClubID: giveup.ClubID,
						}
						mClub, _ := club.GetClubFromDB(giveup.ClubID)
						msg.Diamond = mClub.Diamond //+ r.Cost
						topic.Publish(rs.broker, msg, TopicClubRoomFinish)
					}
				}
				var userNoVoteIds []int32
				for _, userVote := range giveup.GiveupGame.UserStateList {
					if userVote.State > enumr.UserStateDisagree {
						userNoVoteIds = append(userNoVoteIds, userVote.UserID)
					}
				}
				mailReq := &pbmail.SendSysMailRequest{
					MailID: enummail.MailGameGiveUp,
					Ids:    userNoVoteIds,
					Args:   []string{giveup.Password},
				}
				topic.Publish(rs.broker, mailReq, srvmail.TopicSendSysMail)
			}
		}

		if rs.count == 120 {
			mdrList, _ := room.DeadRoomDestroy()
			if mdrList != nil {
				for _, mdr := range mdrList {
					mailReq := &pbmail.SendSysMailRequest{
						MailID: enummail.MailGameOver,
						Ids:    mdr.Ids,
						Args:   []string{mdr.Password},
					}
					mailReq.Ids = append(mailReq.Ids, mdr.GetIdsNotInGame()...)
					if len(mailReq.Ids) > 0 {
						topic.Publish(rs.broker, mailReq, srvmail.TopicSendSysMail)
					}

				}
			}
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
	var r *mdroom.Room
	f := func() error {
		r, err = room.CreateRoom(req.RoomType, req.SubRoomType, req.GameType, req.MaxNumber,
			req.RoundNumber, req.GameParam, req.SettingParam, u, "", 0, 0, req.RoomAdvanceOptions) //, req.JoinType
		if err != nil {
			return err
		}
		return nil
	}
	lock := UserLockKey(u.UserID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s CreateRoom failed: %v", lock, err)
		mderr := errors.Parse(err.Error())
		if mderr.Code == 40054{
			msg := &pbr.RoomNotice{
				Code:    int32(mderr.Code),
				Message: mderr.Detail,
			}
			msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
			topic.Publish(rs.broker, msg, TopicRoomNotice)
		}
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
	err = room.CheckUserInClub(u.UserID, req.ClubID)
	if req.ClubID == 0 && err != nil {
		return errorr.ErrNotClubMember
	}

	if req.VipRoomSettingID > 0 {
		mvrs, err := club.GetVipRoomSettingByID(req.ClubID, req.VipRoomSettingID)
		if err != nil {
			return err
		}
		if mvrs.Status == enumclub.VipRoomSettingStop {
			return errorr.ErrVipRoomStatus
		}
		mdvrs, err := room.GetVipRoomList(req.ClubID, req.VipRoomSettingID)
		if err != nil {
			return err
		}
		if len(mdvrs) > enumr.VipRoomLimit {
			return errorr.ErrVipRoomLimit
		}
		req.RoomType = mvrs.RoomType
		req.SubRoomType = mvrs.SubRoomType
		req.GameType = mvrs.GameType
		req.MaxNumber = mvrs.MaxNumber
		req.RoundNumber = mvrs.RoundNumber
		req.GameParam = mvrs.GameParam
		req.SettingParam = mvrs.SettingParam
		req.RoomAdvanceOptions = mvrs.RoomAdvanceOptions
	}

	var mr *mdroom.Room
	f := func() error {
		mr, err = room.CreateRoom(req.RoomType, req.SubRoomType, req.GameType, req.MaxNumber,
			req.RoundNumber, req.GameParam, req.SettingParam, u, "", req.VipRoomSettingID, req.ClubID, req.RoomAdvanceOptions) //, req.JoinType
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
		//mClub, _ := club.GetClubInfo(mr.ClubID)
		//msg.Diamond = mClub.Diamond
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
		r   *mdroom.Room
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

func (rs *RoomSrv) WatchRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var r *mdroom.Room

	f := func() error {
		r, err = room.WatchRoom(req.Password, u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	//if mr.RoomType == enumr.RoomTypeClub {
	//	lock = ClubJoinRoomLockKey(mr.ClubID)
	//}
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return err
	}
	*rsp = pbr.RoomReply{
		Result: 1,
	}

	//if r.Status == enumr.RoomStatusInit {
	//	msgBack := r.ToProto()
	//	msgBack.UserID = u.UserID
	//	msgBack.CreateOrEnter = enumr.EnterRoom
	//	msgBack.UserRole = enumr.UserRolePlayerBro
	//	if len(r.Users) > 0 {
	//		msgBack.OwnerID = r.Users[0].UserID
	//	} else {
	//		msgBack.OwnerID = u.UserID
	//	}
	//	topic.Publish(rs.broker, msgBack, TopicRoomCreate)
	//} else {
	//	_, err = pubRoomRecovery(rs.broker, u.UserID, req.RoomID, req.RoomType)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (rs *RoomSrv) SitDown(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomUser) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ru       *mdroom.RoomUser
		mdr      *mdroom.Room
		allReady bool
	)
	f := func() error {
		_, ru, mdr, err = room.SitDown(req.Password, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}

	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return err
	}
	rsp = &pbr.RoomUser{UserID: u.UserID}

	//if mdr.Status == enumr.RoomStatusInit {
	//	msgBack := mdr.ToProto()
	//	msgBack.UserID = u.UserID
	//	msgBack.CreateOrEnter = enumr.EnterRoom
	//	msgBack.UserRole = enumr.UserRolePlayerBro
	//	if len(mdr.Users) > 0 {
	//		msgBack.OwnerID = mdr.Users[0].UserID
	//	} else {
	//		msgBack.OwnerID = u.UserID
	//	}
	//	topic.Publish(rs.broker, msgBack, TopicRoomCreate)
	//} else {
	_, err = pubRoomRecovery(rs.broker, u.UserID, req.RoomID, req.RoomType, enumr.RoomSitDown)
	if err != nil {
		return err
	}
	//}
	//
	msgAll := ru.ToProto()
	msgAll.OwnerID = mdr.Users[0].UserID
	msgAll.Ids = mdr.Ids
	msgAll.Ids = append(msgAll.Ids, mdr.GetIdsNotInGame()...)
	for _, ru := range mdr.Users {
		if ru.Role == enumr.UserRoleMaster {
			msgAll.BankerID = ru.UserID
		}
	}
	topic.Publish(rs.broker, msgAll, TopicRoomJoin)
	if mdr.ClubID > 0 {
		msgEnter := &pbr.ClubRoomUser{
			RoomID:           mdr.RoomID,
			UserID:           u.UserID,
			ClubID:           mdr.ClubID,
			Nickname:         u.Nickname,
			Icon:             u.Icon,
			VipRoomSettingID: mdr.VipRoomSettingID,
		}
		topic.Publish(rs.broker, msgEnter, TopicClubRoomJoin)
	}

	msg := rsp
	msg.Ids = mdr.Ids
	topic.Publish(rs.broker, msg, TopicRoomReady)
	if allReady && mdr.Shuffle > 0 {
		msg := &pbr.ShuffleCardBro{
			UserID: mdr.Shuffle,
			Ids:    mdr.Ids,
		}
		topic.Publish(rs.broker, msg, TopicRoomShuffleCardBro)
	}
	return nil
}

func (rs *RoomSrv) EnterRoom(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ru *mdroom.RoomUser
		r  *mdroom.Room
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
	//if mr.RoomType == enumr.RoomTypeClub {
	//	lock = ClubJoinRoomLockKey(mr.ClubID)
	//}
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
	msgAll.Ids = append(msgAll.Ids, mr.GetIdsNotInGame()...)

	for _, ru := range r.Users {
		if ru.Role == enumr.UserRoleMaster {
			msgAll.BankerID = ru.UserID
		}
	}

	topic.Publish(rs.broker, msgAll, TopicRoomJoin)
	if r.ClubID > 0 {
		msgEnter := &pbr.ClubRoomUser{
			RoomID:           r.RoomID,
			UserID:           u.UserID,
			ClubID:           r.ClubID,
			Nickname:         u.Nickname,
			Icon:             u.Icon,
			VipRoomSettingID: r.VipRoomSettingID,
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
		ru *mdroom.RoomUser
		r  *mdroom.Room
	)
	r, err = cacher.GetRoomUserID(u.UserID)
	if err != nil {
		return err
	}
	if r == nil {
		return errorr.ErrUserNotInRoom
	}
	f := func() error {
		ru, r, err = room.LeaveRoom(u)
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
	if ru == nil {
		return nil
	}
	msg := ru.SimplyToProto()
	msg.Ids = r.Ids
	msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
	if r.Status == enumr.RoomStatusDestroy {
		msg.Destroy = 1
	}
	if len(r.Users) > 0 {
		msg.OwnerID = r.Users[0].UserID
		msg.BankerID = msg.OwnerID
	}
	topic.Publish(rs.broker, msg, TopicRoomUnJoin)
	if r.ClubID > 0 {
		msgLeave := &pbr.Room{
			RoomID:           r.RoomID,
			UserID:           u.UserID,
			ClubID:           r.ClubID,
			VipRoomSettingID: r.VipRoomSettingID,
		}

		topic.Publish(rs.broker, msgLeave, TopicClubRoomUnJoin)
		if r.Status == enumr.RoomStatusDestroy {
			msgLeave.UserID = 0
			mClub, _ := club.GetClubInfo(r.ClubID)
			msgLeave.Diamond = mClub.Diamond //+ r.Cost
			topic.Publish(rs.broker, msgLeave, TopicClubRoomFinish)
		}
	}
	if r.Status != enumr.RoomStatusDestroy && (r.GameType == enumr.FourCardGameType || r.GameType == enumr.TwoCardGameType) {
		msgbl := &pbr.UserBankerList{
			BankerIds: r.BankerList,
			Ids:       r.Ids,
		}
		topic.Publish(rs.broker, msgbl, TopicBankerList)
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
		mdr      *mdroom.Room
		allReady bool
	)
	f := func() error {
		allReady, mdr, err = room.GetReady(req.Password, u.UserID, false)
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

	rsp = &pbr.RoomUser{UserID: u.UserID}
	msg := rsp
	msg.Ids = mdr.Ids
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicRoomReady)
	if allReady && mdr.ClubID > 0 && mdr.RoundNow == 1 {
		mClub, _ := club.GetClubInfo(mdr.ClubID)
		msgClub := &pbr.Room{
			ClubID:  mdr.ClubID,
			Diamond: mClub.Diamond,
		}
		topic.Publish(rs.broker, msgClub, TopicClubBalanceUpdate)
	}
	if allReady && mdr.Shuffle > 0 {
		msg := &pbr.ShuffleCardBro{
			UserID: mdr.Shuffle,
			Ids:    mdr.Ids,
		}
		topic.Publish(rs.broker, msg, TopicRoomShuffleCardBro)
	}

	rsp.Ids = nil
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
		result *mdroom.GiveUpGameResult
		mdr    *mdroom.Room
	)
	f := func() error {
		ids, result, mdr, err = room.GiveUpGame(req.Password, u.UserID)
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
	if result == nil {
		return nil
	}
	msg := result.ToProto()
	msg.CountDown = &pbr.CountDown{
		ServerTime: mdr.GiveupAt.Unix(),
		Count:      enumr.RoomGiveupCleanMinutes * 60,
	}
	msg.Ids = ids
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	clubDiamondTopic(rs.broker, mdr)
	if mdr.Status == enumr.RoomStatusGiveUp {
		msg := sendRoomResult(mdr)
		topic.Publish(rs.broker, msg, TopicRoomResult)
	}
	return nil
}

func sendRoomResult(room *mdroom.Room) *pbr.RoomResults {

	roomResults := mdroom.RoomResults{
		RoundNumber: room.RoundNumber,
		RoundNow:    room.RoundNow,
		Status:      room.Status,
		Password:    room.Password,
		List:        room.UserResults,
		RoomType:    room.RoomType,
		SubRoomType: room.SubRoomType,
		CreatedAt:   room.CreatedAt,
	}

	if room.UserResults != nil && len(room.UserResults) > 0 {
		roomResults.List = room.UserResults
	}

	if room.RoomType == enumr.RoomTypeClub && room.SubRoomType == enumr.SubTypeClubMatch {
		for _, ru := range roomResults.List {
			mcm, err := club.GetClubMember(room.ClubID, ru.UserID)
			if err != nil {
				log.Err("reinit get clubmember err:%v", err)
				continue
			}
			ru.ClubCoin = mcm.ClubCoin
		}
	}
	msg := roomResults.ToProto()
	for _, ru := range room.Users {
		mdru := &mdroom.RoomUser{
			UserID:   ru.UserID,
			UserRole: ru.UserRole,
		}
		msg.UserList = append(msg.UserList, mdru.SimplyToProto())
	}
	msg.Ids = room.Ids
	msg.Ids = append(msg.Ids, room.GetIdsNotInGame()...)
	msg.OwnerID = room.Users[0].UserID

	return msg
}

func (rs *RoomSrv) GiveUpVote(ctx context.Context, req *pbr.GiveUpVoteRequest,
	rsp *pbr.GiveUpGameResult) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		ids    []int32
		result *mdroom.GiveUpGameResult
		mdr    *mdroom.Room
	)
	f := func() error {
		ids, result, mdr, err = room.GiveUpVote(req.Password, req.AgreeOrNot, u.UserID)
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
	msg.CountDown = &pbr.CountDown{
		ServerTime: mdr.GiveupAt.Unix(),
		Count:      enumr.RoomGiveupCleanMinutes * 60,
	}
	msg.Ids = ids
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicRoomGiveup)
	if result.Status == enumr.RoomStatusStarted {
		for _, userstate := range result.UserStateList {
			userstate.State = enumr.UserStateWaiting
		}
	}
	clubDiamondTopic(rs.broker, mdr)
	if mdr.Status == enumr.RoomStatusGiveUp {
		msg := sendRoomResult(mdr)
		topic.Publish(rs.broker, msg, TopicRoomResult)
	}
	return nil
}

func clubDiamondTopic(brok broker.Broker, mr *mdroom.Room) {
	if mr.Status == enumr.RoomStatusGiveUp && mr.ClubID > 0 {
		if mr.ClubID > 0 {
			msg := &pbr.Room{
				RoomID:           mr.RoomID,
				ClubID:           mr.ClubID,
				VipRoomSettingID: mr.VipRoomSettingID,
			}
			msg.UserID = 0
			mClub, _ := club.GetClubInfo(mr.ClubID)
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
		mdroom.FeedbackFromProto(req.Feedback))
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
	fb, err := room.CreateFeedback(mdroom.FeedbackFromProto(req), u.UserID, address)
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

func (rs *RoomSrv) GetRoomResultList(ctx context.Context, req *pbr.PageRoomResultListRequest,
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

func (rs *RoomSrv) GetRoomResultByID(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomResults) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	roomresult, err := room.GetRoomResultByID(req.RoomID)
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
	r, err := room.UserRoomCheck(u.UserID)
	if err != nil {
		return err
	}
	msg := &pbr.VoiceChat{
		RoomID:   r.RoomID,
		UserID:   u.UserID,
		FileCode: req.FileCode,
		Ids:      r.Ids,
	}
	msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
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
		r *mdroom.Room
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
	msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
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

func (rs *RoomSrv) GetRoomRecovery(ctx context.Context, req *pbr.Room, rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	rsp.Result = 2
	result, err := pubRoomRecovery(rs.broker, u.UserID, req.RoomID, req.GameType, enumr.RoomRecovery)
	if err != nil {
		return err
	}
	rsp.Result = result
	return nil
}

func pubRoomRecovery(bork broker.Broker, uid int32, rid int32, rtype int32, recoveryType int32) (int32, error) {
	hasRoom, mdroom, err := room.CheckHasRoom(uid)
	if err != nil {
		return 2, err
	}
	if hasRoom || rid > 0 {
		msg := &pbr.RoomExist{}
		msg.UserID = uid
		var gameType int32
		if hasRoom {
			msg.RoomID = mdroom.RoomID
			gameType = mdroom.GameType
		} else {
			msg.RoomID = rid
			gameType = rtype
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
		case enumr.FourCardGameType:
			top = TopicRoomFourCardExist
			break
		case enumr.TwoCardGameType:
			top = TopicRoomTwoCardExist
			break
		case enumr.RunCardGameType:
			top = TopicRoomRunCardExist
			break
		}
		msg.RecoveryType = recoveryType
		topic.Publish(bork, msg, top)
		return 1, nil
	}
	return 2, nil
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
	//if mr.GameType == enumr.ThirteenGameType {
	//	return errorr.ErrGameType
	//}
	var mdr *mdroom.Room
	f := func() error {
		mdr, err = room.GameStart(u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s game start failed: %v", lock, err)
		mderr := errors.Parse(err.Error())
		if mderr.Code == 40054{
			msg := &pbr.RoomNotice{
				Code:    int32(mderr.Code),
				Message: mderr.Detail,
			}
			msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
			topic.Publish(rs.broker, msg, TopicRoomNotice)
		}
		return err
	}
	msg := &pbr.Room{
		Password: mr.Password,
		ClubID:   mr.ClubID,
	}
	if mdr.ClubID > 0 {
		mClub, _ := club.GetClubInfo(mdr.ClubID)
		msgClub := &pbr.Room{
			ClubID:  mdr.ClubID,
			Diamond: mClub.Diamond,
		}
		topic.Publish(rs.broker, msgClub, TopicClubBalanceUpdate)
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

func (rs *RoomSrv) ShuffleCard(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var (
		mdr      *mdroom.Room
		allReady bool
	)
	f := func() error {
		allReady, mdr, err = room.GetReady(req.Password, u.UserID, true)
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
	msg := &pbr.RoomUser{UserID: u.UserID}
	msg.Ids = mdr.Ids
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicRoomReady)
	if allReady && mdr.Shuffle > 0 {
		msg := &pbr.ShuffleCardBro{
			UserID: mdr.Shuffle,
			Ids:    mdr.Ids,
		}
		msg.Ids = append(msg.Ids, mdr.WatchIds...)
		topic.Publish(rs.broker, msg, TopicRoomShuffleCardBro)
	}
	rsp.Result = 1
	return nil
}

func (rs *RoomSrv) RoomChat(ctx context.Context, req *pbr.RoomChatRequest, rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	r, err := room.UserRoomCheck(u.UserID)
	if err != nil {
		return err
	}
	msg := &pbr.RoomChat{
		RoomID:   r.RoomID,
		UserID:   u.UserID,
		ChatCode: req.ChatCode,
		Ids:      r.Ids,
	}
	msg.Ids = append(msg.Ids, r.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicRoomChat)
	return nil
}

func (rs *RoomSrv) PageSpecialGameList(ctx context.Context,
	req *pbr.PageSpecialGameListRequest, rsp *pbr.PageSpecialGameListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := room.PageSpecialGameList(page,
		mdroom.GameRecordFromProto(req.GameRecord))
	if err != nil {
		return err
	}
	for _, co := range l {
		rsp.List = append(rsp.List, co.ToProto())
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (rs *RoomSrv) SetBankerList(ctx context.Context, req *pbr.SetBankerListRequest,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdr := &mdroom.Room{}
	f := func() error {
		mdr, err = room.SetBankerList(req.Password, u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)

	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set banker list failed: %v", lock, err)
		return err
	}
	//*rsp = pbr.RoomReply{
	//	Result: 1,
	//}
	rsp.Result = 1
	msg := &pbr.UserBankerList{
		BankerIds: mdr.BankerList,
		Ids:       mdr.Ids,
	}
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicBankerList)
	return nil
}

func (rs *RoomSrv) OutBankerList(ctx context.Context, req *pbr.OutBankerListRequest,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdr := &mdroom.Room{}
	f := func() error {
		mdr, err = room.OutBankerList(req.Password, u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)

	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s out banker list failed: %v", lock, err)
		return err
	}
	//*rsp = pbr.RoomReply{
	//	Result: 1,
	//}
	rsp.Result = 1
	msg := &pbr.UserBankerList{
		BankerIds: mdr.BankerList,
		Ids:       mdr.Ids,
	}
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicBankerList)
	return nil
}

func (rs *RoomSrv) GetVipRoomList(ctx context.Context, req *pbr.Room,
	rsp *pbr.GetVipRoomListReply) error {
	rsp.Result = 2
	if req.ClubID == 0 || req.VipRoomSettingID == 0 {
		return errorr.ErrIDZero
	}
	list, err := room.GetVipRoomList(req.ClubID, req.VipRoomSettingID)
	if err != nil {
		return err
	}
	utilpb.ProtoSlice(list, &rsp.List)
	rsp.Result = 1
	return nil
}

func (rs *RoomSrv) GetRoomRoundNow(ctx context.Context, req *pbr.Room,
	rsp *pbr.ClubRoomLogReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	list, err := room.GetRoomRoundNow(req.GameType)
	if err != nil {
		return err
	}
	utilpb.ProtoSlice(list, &rsp.List)
	return nil
}

func (rs *RoomSrv) SetSuspend(ctx context.Context, req *pbr.Room,
	rsp *pbr.ClubRoomLogReply) error {
	//u, err := auth.GetUser(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//var mdr      *mdroom.Room
	//f := func() error {
	//	mdr, err = room.SetSuspend(u.UserID)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//lock := RoomLockKey(req.Password)
	//err = gsync.GlobalTransaction(lock, f)
	//if err != nil {
	//	log.Err("%s set suspend room failed: %v", lock, err)
	//	return err
	//}
	//
	//rsp = &pbr.RoomUser{UserID: u.UserID}
	//msg := rsp
	//msg.Ids = mdr.Ids
	//msg.Ids = append(msg.Ids, mdr.WatchIds...)
	//topic.Publish(rs.broker, msg, TopicRoomReady)
	//if allReady && mdr.ClubID > 0 && mdr.RoundNow == 1{
	//	mClub, _ := club.GetClubInfo(mdr.ClubID)
	//	msgClub := &pbr.Room{
	//		ClubID:  mdr.ClubID,
	//		Diamond: mClub.Diamond,
	//	}
	//	topic.Publish(rs.broker, msgClub, TopicClubBalanceUpdate)
	//}
	//if allReady && mdr.Shuffle > 0 {
	//	msg := &pbr.ShuffleCardBro{
	//		UserID: mdr.Shuffle,
	//		Ids:    mdr.Ids,
	//	}
	//	topic.Publish(rs.broker, msg, TopicRoomShuffleCardBro)
	//}
	//
	//rsp.Ids = nil
	return nil
}

func (rs *RoomSrv) SetRestore(ctx context.Context, req *pbr.Room,
	rsp *pbr.RoomReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	var (
		mdr *mdroom.Room
		ru  *mdroom.RoomUser
	)
	f := func() error {
		mdr, ru, err = room.SetRestore(u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := RoomLockKey(req.Password)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s set suspend room failed: %v", lock, err)
		return err
	}
	rsp.Result = 1

	msg := &pbr.RoomUser{
		UserID:   u.UserID,
		UserRole: ru.UserRole,
		Ids:      mdr.Ids,
	}
	msg.Ids = append(msg.Ids, mdr.GetIdsNotInGame()...)
	topic.Publish(rs.broker, msg, TopicUserRestore)
	return nil
}

func (rs *RoomSrv) GetRoomListByIds(ctx context.Context, req *pbr.GetRoomListByIDSRequest,
	rsp *pbr.GetRoomListByIDSReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	pbrs := room.GetRoomListByIDS(req.Ids)
	if err != nil {
		return err
	}
	rsp.List = pbrs
	rsp.Result = 1
	return nil
}
