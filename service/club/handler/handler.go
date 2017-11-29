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
	mduser "playcards/model/user/mod"
	pbclub "playcards/proto/club"
	"playcards/utils/auth"
	"playcards/utils/log"
	utilproto "playcards/utils/proto"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"time"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

type ClubSrv struct {
	server server.Server
	broker broker.Broker
}

func ClubLockKey(clubid int32, optype string) string {
	return fmt.Sprintf("playcards.club.op.lock:%d|%s", clubid, optype)
}

func ClubRoomLockKey(clubid int32) string {
	return fmt.Sprintf("playcards.club.op.lock:%d", clubid)
}

func NewHandler(s server.Server) *ClubSrv {
	cs := &ClubSrv{
		server: s,
		broker: s.Options().Broker,
	}
	cs.Init()
	return cs
}

func (cs *ClubSrv) Init() {
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
	err := club.CreateClub(req.ClubName, req.CreatorID, req.CreatorProxy)
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
	f := func() error {
		err := club.UpdateClub(mclub)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID, enumclub.ClubOpUpdateRoom)
	err := gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club update failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) SetClubRoomFlag(ctx context.Context,
	req *pbclub.SetClubRoomFlagRequest, rsp *pbclub.ClubReply) error {
	err := club.SetClubRoomFlag(req.ClubID, req.RoomID)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) RemoveClubMember(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {

	f := func() error {
		err := club.RemoveClubMember(req.ClubID, req.UserID, enumclub.ClubMemberStatusRemoved)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID, enumclub.ClubOpRemoveMember)
	err := gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s remove club member failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID: req.ClubID,
		UserID: req.UserID,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberLeave)
	return nil
}

func (cs *ClubSrv) CreateClubMember(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	var (
		user *mduser.User
		err  error
	)
	f := func() error {
		user, err = club.CreateClubMember(req.ClubID, req.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubLockKey(req.ClubID, enumclub.ClubOpCreateMember)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s create club member failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID:   req.ClubID,
		UserID:   user.UserID,
		Nickname: user.Nickname,
		Icon:     user.Icon,
		Online:   cacheuser.GetUserOnlineStatus(user.UserID),
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberJoin)
	return nil
}

func (cs *ClubSrv) JoinClub(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = club.JoinClub(req.ClubID, u)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	return nil
}

func (cs *ClubSrv) LeaveClub(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
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
	lock := ClubLockKey(req.ClubID, enumclub.ClubOpLeaveClub)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s create club leave  failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID: req.ClubID,
		UserID: u.UserID,
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
	ci, err := club.GetClub(u)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	topic.Publish(cs.broker, ci, TopicClubInfo)
	return nil
}

func (cs *ClubSrv) PageClub(ctx context.Context,
	req *pbclub.PageClubRequest, rsp *pbclub.PageClubReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClub(page,
		mdclub.ClubFromProto(req.Club))
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
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := club.PageClubMember(page,
		mdclub.ClubMemberFromProto(req.ClubMember))
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

func (cs *ClubSrv) PageClubRoom(ctx context.Context,
	req *pbclub.PageClubRoomRequest, rsp *pbclub.PageClubRoomReply) error {
	rsp.Result = 2
	l, err := club.PageClubRoom(req.ClubID, req.Page, req.PageSize,req.Flag)
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
	req *pbclub.ClubRechargeRequest, rsp *pbclub.ClubReply) error {
	if req.Amount < 0 {
		return errorclub.ErrClubRecharge
	}
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	f := func() error {
		err = club.SetClubBalance(req.Amount, enumclub.TypeDiamond, req.ClubID,
			enumbill.JournalTypeClubRecharge, time.Now().Unix(), int64(u.UserID))
		if err != nil {
			return err
		}
		return nil
	}
	lock := ClubRoomLockKey(req.ClubID)
	err = gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s club recharge failed: %v", lock, err)
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	rsp.Result = 1
	return nil
}

func (cs *ClubSrv) SetBlackList(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = club.SetBlackList(req.ClubID, req.UserID, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID: req.ClubID,
		UserID: req.UserID,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberLeave)
	return nil
}

func (cs *ClubSrv) UpdateClubExamine(ctx context.Context,
	req *pbclub.ClubMember, rsp *pbclub.ClubReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = club.UpdateClubExamine(req.ClubID, req.UserID, req.Status, u.UserID)
	if err != nil {
		return err
	}
	*rsp = pbclub.ClubReply{
		Result: 1,
	}
	msgAll := &pbclub.ClubMember{
		ClubID: req.ClubID,
		UserID: req.UserID,
	}
	topic.Publish(cs.broker, msgAll, TopicClubMemberJoin)
	return nil
}
