package handler

import (
	"fmt"
	enumbill "playcards/model/bill/enum"
	"playcards/model/club"
	enumclub "playcards/model/club/enum"
	errorclub "playcards/model/club/errors"
	mdclub "playcards/model/club/mod"
	mdpage "playcards/model/page"
	cacheuser "playcards/model/user/cache"
	cacheclub "playcards/model/club/cache"
	mduser "playcards/model/user/mod"
	pbclub "playcards/proto/club"
	pbroom "playcards/proto/room"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilproto "playcards/utils/proto"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	pbmail "playcards/proto/mail"
	srvmail "playcards/service/mail/handler"
	utilpb "playcards/utils/proto"
	enumcon "playcards/model/common/enum"
	cachelog "playcards/model/log/cache"
	"time"
	"playcards/model/room"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

type ClubSrv struct {
	server server.Server
	broker broker.Broker
}

//func ClubLockKey(clubid int32, optype string) string {
//	return fmt.Sprintf("playcards.club.op.lock:%d|%s", clubid, optype)
//}

func ClubLockKey(clubid int32) string {
	return fmt.Sprintf("playcards.club.op.lock:%d", clubid)
}

func NewHandler(s server.Server) *ClubSrv {
	cs := &ClubSrv{
		server: s,
		broker: s.Options().Broker,
	}
	cs.init()
	return cs
}

func (cs *ClubSrv) init() {
	club.RefreshAllFromDB()
}

func (cs *ClubSrv) RefreshAllFromDB(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubReply) error {
	err := club.RefreshAllFromDB()
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) CreateClub(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	//if u.Rights != auth.RightsAdmin {
	//
	//	return errorclub.ErrNotCreatorID
	//}
	if u.Rights != auth.RightsAdmin || req.CreatorID == 0 {
		req.CreatorID = u.UserID
		req.CreatorProxy = u.ProxyID
	}
	err = club.CreateClub(req.ClubName, req.CreatorID, req.CreatorProxy)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) UpdateClub(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubReply) error {
	mclub := mdclub.ClubFromProto(req)
	if mclub == nil {
		return errorclub.ErrUpdateClubNull
	}
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	f := func() error {
		//fmt.Printf("AAAAAUpdateClub:%+v|%+v\n", mclub,mclub.Setting)
		if mclub.Setting != nil {
			if mclub.Setting.UserCreateRoom == 100 || mclub.Setting.UserCreateRoom == 200 {
				var userCreateRoom int32 = mclub.Setting.UserCreateRoom / 100
				mclub.Setting = mdClub.Setting
				mclub.Setting.UserCreateRoom = userCreateRoom
			} else if mclub.Setting.CostType == 0 || mclub.Setting.CostRange == 0 ||
				mclub.Setting.UserCreateRoom == 0 {
				mclub.Setting = mdClub.Setting
			}
		}
		//fmt.Printf("BBBBBBUpdateClub:%+v\n", mclub.Setting)
		err := club.UpdateClub(mclub)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s club update failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msg := mclub.ToProto()
	topic.Publish(cs.broker, msg, TopicUpdateClub)
	return nil
}

func (cs *ClubSrv) SetClubRoomFlag(ctx context.Context,
	req *pbclub.SetClubRoomFlagRequest, rsp *pbclub.SetClubRoomFlagReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	err = club.SetClubRoomFlag(req.ClubID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = pbclub.SetClubRoomFlagReply{
		ClubID: req.ClubID,
		RoomID: req.RoomID,
	}
	return nil
}

func (cs *ClubSrv) RemoveClubMember(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	f := func() error {
		err := club.RemoveClubMember(req.ClubID, req.UserID, enumclub.ClubMemberStatusRemoved)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(mdClub.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s remove club member failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	memberCount, onlineCount := cacheclub.GetMemberCount(req.ClubID)
	msgAll := &pbclub.ClubMember{
		ClubID:       req.ClubID,
		UserID:       req.UserID,
		MemberCount:  memberCount,
		MemberOnline: onlineCount,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberLeave)

	mailReq := &pbmail.SendSysMailRequest{
		MailID: enumclub.MailClubUnJoin,
		Ids:    []int32{req.UserID},
		Args:   []string{mdClub.ClubName},
	}
	topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	return nil
}

func (cs *ClubSrv) CreateClubMember(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {

	//u, err := auth.GetUser(ctx)
	//if err != nil {
	//	return err
	//}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	//if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
	//	return errorclub.ErrNotCreatorID
	//}
	var user *mduser.User
	f := func() error {
		mdClub, user, err = club.CreateClubMember(req.ClubID, req.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s create club member failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID:    req.ClubID,
		UserID:    user.UserID,
		Icon:      user.Icon,
		CreatorID: mdClub.CreatorID,
		Online:    cacheuser.GetUserOnlineStatus(user.UserID),
	}
	_, creator := cacheuser.GetUserByID(mdClub.CreatorID)
	if creator != nil {
		msgAll.CreatorName = creator.Nickname
		msgAll.Icon = creator.Icon
	}

	topic.Publish(cs.broker, msgAll, TopicClubMemberJoin)

	mailReq := &pbmail.SendSysMailRequest{
		MailID: enumclub.MailClubJoin,
		Ids:    []int32{req.UserID},
		Args:   []string{mdClub.ClubName},
	}
	topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	return nil
}

func (cs *ClubSrv) JoinClub(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	f := func() error {
		err = club.JoinClub(mdClub.ClubID, u)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s club update failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	//mailReq := &pbmail.SendSysMailRequest{
	//	MailID: enumclub.MailClubJoin,
	//	Ids:    []int32{u.UserID},
	//	Args:   []string{mdClub.ClubName},
	//}
	//topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	return nil
}

func (cs *ClubSrv) LeaveClub(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.Club) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	f := func() error {
		err = club.RemoveClubMember(req.ClubID, u.UserID, enumclub.ClubMemberStatusLeave)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(mdClub.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s create club leave  failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.Club{
		ClubID: req.ClubID,
	}
	memberCount, onlineCount := cacheclub.GetMemberCount(req.ClubID)
	msgAll := &pbclub.ClubMember{
		ClubID:       req.ClubID,
		UserID:       u.UserID,
		MemberCount:  memberCount,
		MemberOnline: onlineCount,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberLeave)

	return nil
}

func (cs *ClubSrv) GetClub(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	if req.ClubID == 0 {
		req.ClubID = u.ClubID
	}

	//if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
	//	return errorclub.ErrNotCreatorID
	//}
	ci, err := club.GetClub(req.ClubID, u.UserID, true) //, u.ClubID
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	topic.Publish(cs.broker, ci, TopicClubInfo)
	return nil
}

func (cs *ClubSrv) GetClubByClubID(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubInfo) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	_, err = cacheclub.GetClubMember(req.ClubID, u.UserID)
	if err != nil {
		return err
	}

	//mdClub, err := cacheclub.GetClub(req.ClubID)
	//if err != nil {
	//	return err
	//}
	//
	//if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
	//	return errorclub.ErrNotCreatorID
	//}

	ci, err := club.GetClub(req.ClubID, u.UserID, false) //, u.ClubID
	if err != nil {
		return err
	}
	if req.ClubID > 0 {
		u.ClubID = req.ClubID
		cacheuser.SimpleUpdateUser(u)
	}
	*rsp = *ci
	//topic.Publish(cs.broker, ci, TopicClubInfo)
	return nil
}

func (cs *ClubSrv) PageClub(ctx context.Context,
	req *pbclub.PageClubRequest, rsp *pbclub.PageClubReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClub(page, mdclub.ClubFromProto(req.Club))
	if err != nil {
		return err
	}
	err = utilproto.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) PageClubMember(ctx context.Context,
	req *pbclub.PageClubMemberRequest, rsp *pbclub.PageClubMemberReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin {
		mdClub, err := cacheclub.GetClub(req.ClubMember.ClubID)
		if err != nil {
			return err
		}
		if mdClub.CreatorID != u.UserID {
			return errorclub.ErrNotCreatorID
		}
	}

	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClubMember(page,
		mdclub.ClubMemberFromProto(req.ClubMember))
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	rsp.List = l
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) PageClubRoom(ctx context.Context,
	req *pbclub.PageClubRoomRequest, rsp *pbclub.PageClubRoomReply) error {
	rsp.Result = 2
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	l, err := club.PageClubRoom(req.ClubID, req.Page, req.PageSize, req.Flag)
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	rsp.List = l
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) ClubRecharge(ctx context.Context,
	req *pbclub.ClubRechargeRequest, rsp *pbclub.ClubRechargeReply) error {
	if req.Amount < 0 {
		return errorclub.ErrClubRecharge
	}
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	f := func() error {
		err = club.SetClubBalance(req.Amount, req.AmountType, req.ClubID,
			enumbill.JournalTypeClubRecharge, time.Now().Unix(), int64(u.UserID), u.Rights == auth.RightsAdmin)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		cachelog.SetErrLog(enumclub.ServiceCode, err.Error())
		log.Err("%s club recharge failed: %v", lock, err)
		return err
	}
	mdClub, err = cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	rsp.ClubID = req.ClubID
	rsp.Amount = mdClub.Diamond
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) SetBlackList(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubMember) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	f := func() error {
		err = club.SetBlackList(req.ClubID, req.UserID, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(mdClub.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club set black list failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubMember{
		ClubID: req.ClubID,
	}
	msgAll := &pbclub.ClubMember{
		ClubID: req.ClubID,
		UserID: req.UserID,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberLeave)

	mailReq := &pbmail.SendSysMailRequest{
		MailID: enumclub.MailClubUnJoin,
		Ids:    []int32{req.UserID},
		Args:   []string{mdClub.ClubName},
	}
	topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)

	return nil
}

func (cs *ClubSrv) CancelBlackList(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubMember) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}

	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	f := func() error {
		err = club.CancelBlackList(req.ClubID, req.UserID, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(mdClub.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club set black list failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubMember{
		ClubID: req.ClubID,
	}
	return nil
}

func (cs *ClubSrv) PageBlackListMember(ctx context.Context,
	req *pbclub.PageClubMemberRequest, rsp *pbclub.PageClubMemberReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubMember.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageBlackListMember(page, req.ClubMember.ClubID)
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("PageBlackListMember:%+v|%+v\n",req,l)
	rsp.List = l
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) PageClubExamineMember(ctx context.Context,
	req *pbclub.PageClubMemberRequest, rsp *pbclub.PageClubMemberReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubMember.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClubExamineMember(page, req.ClubMember.ClubID)
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	rsp.List = l
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

//func (cs *ClubSrv) CreateClubExamine(ctx context.Context,
//	req *pbclub.ClubMember, rsp *pbclub.ClubMember) error {
//	u, err := auth.GetUser(ctx)
//	if err != nil {
//		return err
//	}
//	mdClub, err := cacheclub.GetClub(req.ClubID)
//	if err != nil {
//		return err
//	}
//	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
//		req.UserID = u.UserID
//	}
//	err = club.CreateClubExamine(req.ClubID, req.UserID, u.UserID)
//	if err != nil {
//		return err
//	}
//	*rsp = pbclub.ClubMember{
//		ClubID: 1,
//	}
//	return nil
//}

func (cs *ClubSrv) UpdateClubExamine(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubMember) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	//mdClub := &mdclub.Club{}
	f := func() error {
		_, err = club.UpdateClubExamine(req.ClubID, req.UserID, req.Status, u.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club set examine failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubMember{
		ClubID: req.ClubID,
	}
	if req.Status == enumcon.ExamineStatusPass {
		_, muser := cacheuser.GetUserByID(req.UserID)
		msgAll := &pbclub.ClubMember{
			ClubID:    req.ClubID,
			UserID:    muser.UserID,
			Nickname:  muser.Nickname,
			Online:    cacheuser.GetUserOnlineStatus(muser.UserID),
			CreatorID: mdClub.CreatorID,
		}
		msgAll.ClubName = mdClub.ClubName
		_, creator := cacheuser.GetUserByID(mdClub.CreatorID)
		if creator != nil {
			msgAll.CreatorName = creator.Nickname
			msgAll.Icon = creator.Icon
		}
		topic.Publish(cs.broker, msgAll, TopicClubMemberJoin)

		mailReq := &pbmail.SendSysMailRequest{
			MailID: enumclub.MailClubJoin,
			Ids:    []int32{req.UserID},
			Args:   []string{mdClub.ClubName},
		}
		topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	} else if req.Status == enumcon.ExamineStatusRefuse {
		mailReq := &pbmail.SendSysMailRequest{
			MailID: enumclub.MailClubUnRefuse,
			Ids:    []int32{req.UserID},
			Args:   []string{mdClub.ClubName},
		}
		topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	}

	return nil
}

func (cs *ClubSrv) AddClubMemberClubCoin(ctx context.Context,
	req *pbclub.GainClubCoinRequest, rsp *pbclub.GainClubCoinReply) error {
	if req.Amount < 0 {
		return errorclub.ErrClubRecharge
	}

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	var amount int64
	f := func() error {
		amount, err = club.AddClubMemberClubCoin(req.ClubID, req.UserID, req.Amount)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club recharge failed: %v", lock, err)
		return err
	}
	msg := &pbclub.GainClubCoinReply{
		Result:   1,
		ClubCoin: amount,
		UserID:   req.UserID,
		ClubID:   req.ClubID,
	}
	*rsp = *msg
	topic.Publish(cs.broker, msg, TopicAddClubCoin)
	return nil
}

func (cs *ClubSrv) ClubMemberOfferUpClubCoin(ctx context.Context,
	req *pbclub.GainClubCoinRequest, rsp *pbclub.GainClubCoinReply) error {
	if req.Amount < 0 {
		return errorclub.ErrClubRecharge
	}
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	var amount int64
	f := func() error {
		amount, err = club.ClubMemberOfferUpClubCoin(req.ClubID, u.UserID, req.Amount)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club recharge failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.GainClubCoinReply{
		Result:   1,
		ClubCoin: amount,
	}
	return nil
}

func (cs *ClubSrv) PageClubJournal(ctx context.Context,
	req *pbclub.PageClubJournalRequest, rsp *pbclub.PageClubJournalReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClubJournal(page, req.ClubID, req.Status)
	if err != nil {
		return err
	}
	err = utilproto.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) PageClubMemberJournal(ctx context.Context,
	req *pbclub.PageClubJournalRequest, rsp *pbclub.PageClubJournalReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	l, rows, err := club.PageClubMemberJournal(page, u.UserID, req.ClubID)
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	rsp.List = l
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) UpdateClubJournal(ctx context.Context,
	req *pbclub.UpdateClubJournalRequest, rsp *pbclub.ClubReply) error {

	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	if req.ClubJournalID == 0 {
		return errorclub.ErrIDZero
	}
	err = club.UpdateClubJournal(req.ClubJournalID, req.ClubID)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) UpdateClubMemberStatus(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	err = club.UpdateClubMemberStatus(req.ClubID, req.UserID, req.Status)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) GetClubMemberCoinRank(ctx context.Context,
	req *pbclub.GetClubMemberCoinRankRequest, rsp *pbclub.PageClubMemberReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	//u, err := auth.GetUser(ctx)
	//if err != nil {
	//	return err
	//}
	//if req.ClubID == 0 {
	//	return errorclub.ErrClubIDZero
	//}
	//mdClub, err := cacheclub.GetClub(req.ClubID)
	//if err != nil {
	//	return err
	//}
	//if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
	//	return errorclub.ErrNotCreatorID
	//}
	l, rows, err := club.GetClubMemberCoinRank(page, req.ClubID)
	if err != nil {
		return err
	}
	//err = utilproto.ProtoSlice(l, &rsp.List)
	//if err != nil {
	//	return err
	//}
	rsp.List = l
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) UpdateClubProxyID(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.ProxyID > 0 {
		return errorclub.ErrAlreadyHasProxyID
	}
	err = club.UpdateClubProxyID(req.CreatorID, req.CreatorProxy)
	if err != nil {
		return err
	}

	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) GetClubsByMemberID(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.GetClubsByMemberIDReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	pbcs, err := club.GetClubsByMemberID(u.UserID)
	if err != nil {
		return err
	}

	rsp.List = pbcs
	return nil
}

func (cs *ClubSrv) CreateVipRoomSetting(ctx context.Context,
	req *pbclub.VipRoomSetting, rsp *pbclub.VipRoomSetting) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	mdVrs := mdclub.VipRoomSettingFromProto(req)
	err = club.CreateVipRoomSetting(u, mdVrs)
	if err != nil {
		return err
	}
	//*rsp = pbclub.ClubReply{
	//	Result: 1,
	//}
	msg := mdVrs.ToProto()
	*rsp = *msg
	msg.Status = enumclub.VipRoomSettingNew
	topic.Publish(cs.broker, msg, TopicUpdateVipRoomSetting)
	rsp.ClubID = mdVrs.ClubID
	return nil
}

func (cs *ClubSrv) UpdateVipRoomSetting(ctx context.Context,
	req *pbclub.VipRoomSetting, rsp *pbclub.VipRoomSetting) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	mvrs := mdclub.VipRoomSettingFromProto(req)
	if mvrs.ID == 0 {
		return errorclub.ErrVipRoomSettingNoFind
	}
	f := func() error {
		err := club.UpdateVipRoomSetting(mvrs)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s vip room setting update failed: %v", lock, err)
		return err
	}
	msg := mvrs.ToProto()
	*rsp = *msg
	topic.Publish(cs.broker, msg, TopicUpdateVipRoomSetting)
	return nil
}

func (cs *ClubSrv) UpdateVipRoomSettingStatus(ctx context.Context,
	req *pbclub.VipRoomSetting, rsp *pbclub.VipRoomSetting) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	mvrs := &mdclub.VipRoomSetting{
		ClubID: req.ClubID,
		ID:     req.VipRoomSettingID,
		Status: req.Status,
	}
	f := func() error {
		err := club.UpdateVipRoomSettingStatus(mvrs)
		if err != nil {
			return err
		}
		return nil
	}

	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s vip room setting update failed: %v", lock, err)
		return err
	}
	msg := mvrs.ToProto()
	*rsp = *msg
	topic.Publish(cs.broker, msg, TopicUpdateVipRoomSetting)
	return nil
}

func (cs *ClubSrv) GetVipRoomSettingList(ctx context.Context,
	req *pbclub.GetVipRoomSettingListRequest, rsp *pbclub.GetVipRoomSettingListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin {
		_, err := cacheclub.GetClubMember(req.ClubID, u.UserID)
		if err != nil {
			return err
		}
		//return errorclub.ErrNotCreatorID
	}
	list, err := club.GetVipRoomSettingList(req.ClubID)
	err = utilpb.ProtoSlice(list, &rsp.List)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ClubSrv) ClubReName(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.Club) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if len(req.ClubName) == 0 {
		return errorclub.ErrNameLen
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}

	f := func() error {
		mc := &mdclub.Club{ClubID: req.ClubID, ClubName: req.ClubName}
		err := club.UpdateClub(mc)
		if err != nil {
			return err
		}
		mdClub.ClubName = req.ClubName
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club update name failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.Club{
		ClubID:   req.ClubID,
		ClubName: req.ClubName,
	}
	return nil
}

func (cs *ClubSrv) ClubDelete(ctx context.Context,
	req *pbclub.Club, rsp *pbclub.ClubDeleteReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	var ids []int32
	var onLineIds []int32
	f := func() error {
		ids, onLineIds, err = club.DeleteClub(req.ClubID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club delete failed: %v", lock, err)
		return err
	}
	temp := &pbclub.Club{
		ClubID: req.ClubID,
	}
	msg := &pbclub.ClubDeleteReply{
		Club:      temp,
		Ids:       ids,
		OnlineIds: onLineIds,
	}
	*rsp = *msg
	topic.Publish(cs.broker, msg, TopicClubDelete)

	mailReq := &pbmail.SendSysMailRequest{
		MailID: enumclub.MailClubDelete,
		Ids:    msg.Ids,
		Args:   []string{mdClub.ClubName},
	}
	topic.Publish(cs.broker, mailReq, srvmail.TopicSendSysMail)
	return nil
}

func (cs *ClubSrv) GetClubRoomLog(ctx context.Context,
	req *pbclub.ClubRoomLogRequest, rsp *pbclub.ClubRoomLogReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	list, err := club.GetClubRoomLog(req.ClubID)
	if err != nil {
		return err
	}

	rsp.ClubID = req.ClubID
	err = utilpb.ProtoSlice(list, &rsp.List)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ClubSrv) PageClubRoomResultList(ctx context.Context, req *pbclub.PageClubRoomResultListRequest,
	rsp *pbroom.RoomResultListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	mdClub, err := cacheclub.GetClub(req.ClubID)
	if err != nil {
		return err
	}
	if u.Rights != auth.RightsAdmin && mdClub.CreatorID != u.UserID {
		return errorclub.ErrNotCreatorID
	}
	page := mdpage.PageOptionFromProto(req.Page)
	roomresult, err := room.PageClubRoomResultList(page, req.ClubID)
	if err != nil {
		return err
	}
	*rsp = *roomresult
	return nil
}

func (cs *ClubSrv) PageClubMemberRoomResultList(ctx context.Context, req *pbclub.PageClubMemberRoomResultListRequest,
	rsp *pbroom.RoomResultListReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	page := mdpage.PageOptionFromProto(req.Page)
	roomresult, err := room.PageClubMemberRoomResultList(page, req.ClubID, u.UserID)
	if err != nil {
		return err
	}
	*rsp = *roomresult
	return nil
}
