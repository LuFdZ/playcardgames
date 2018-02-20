package handler

import (
	"playcards/model/mail"
	mdgame "playcards/model/mail/mod"
	pbgame "playcards/proto/mail"
	enumgame "playcards/model/mail/enum"
	mdpage "playcards/model/page"
	gsync "playcards/utils/sync"
	"playcards/utils/topic"
	"playcards/utils/auth"
	"time"

	"golang.org/x/net/context"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
)

type MailSrv struct {
	server server.Server
	broker broker.Broker
	count  int32
}

func NewHandler(s server.Server, gt *gsync.GlobalTimer) *MailSrv {
	m := &MailSrv{
		server: s,
		broker: s.Options().Broker,
	}
	m.init()
	m.update(gt)
	return m
}

func (ms *MailSrv) init() {
	mail.RefreshAllMailInfoFromDB()
}

func (ms *MailSrv) update(gt *gsync.GlobalTimer) {
	lock := "playcards.mail.update.lock"
	f := func() error {
		mail.CleanOverdueByCreateAt()
		return nil
	}
	gt.Register(lock, time.Minute*enumgame.LoopTime, f)
}

func (ms *MailSrv) CreateMailInfo(ctx context.Context,
	req *pbgame.CreateMailInfoRequest, rsp *pbgame.DefaultReply) error {
	mi := mdgame.MailInfoFromProto(req.MailInfo)
	if req.ItemModelA != nil {
		mi.ItemModeList = append(mi.ItemModeList, mdgame.ItemModelFromProto(req.ItemModelA))
	}
	if req.ItemModelB != nil {
		mi.ItemModeList = append(mi.ItemModeList, mdgame.ItemModelFromProto(req.ItemModelB))
	}
	if req.ItemModelC != nil {
		mi.ItemModeList = append(mi.ItemModeList, mdgame.ItemModelFromProto(req.ItemModelC))
	}

	_, err := mail.CreateMailInfo(mi)
	if err != nil {
		return err
	}

	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}
	*rsp = *reply
	return nil
}

func (ms *MailSrv) UpdateMailInfo(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	mi := mdgame.MailInfoFromProto(req)
	_, err := mail.UpdateMailInfo(mi)
	if err != nil {
		return err
	}

	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}

	*rsp = *reply
	return nil
}

func (ms *MailSrv) SendSysMail(ctx context.Context, req *pbgame.SendSysMailRequest, rsp *pbgame.DefaultReply) error {
	_, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = mail.SendSysMail(req.MailID, req.Ids, req.Args)
	if err != nil {
		return err
	}
	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}
	*rsp = *reply
	return nil
}

func (ms *MailSrv) SendMail(ctx context.Context, req *pbgame.SendMailRequest,
	rsp *pbgame.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	msl, err := mail.SendMail(u.UserID, mdgame.SendMailLogFromProto(req.MailSend), req.Ids, req.Channel)
	if err != nil {
		return err
	}
	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
	}

	msg := &pbgame.NewMailNoticeBro{
		Context: &pbgame.MailSendLog{SendID: msl.SendID},
		Ids:     req.Ids,
	}
	topic.Publish(ms.broker, msg, TopicMailNotice)
	*rsp = *reply
	return nil
}

func (ms *MailSrv) RefreshAllMailInfoFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllMailInfoFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}

func (ms *MailSrv) RefreshAllMailSendLogFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllSendMailLogFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}

func (ms *MailSrv) RefreshAllPlayerMailFromDB(ctx context.Context,
	req *pbgame.MailInfo, rsp *pbgame.DefaultReply) error {
	err := mail.RefreshAllPlayerMailFromDB()
	if err != nil {
		return err
	}
	*rsp = pbgame.DefaultReply{
		Result: 1,
	}
	return nil
}

func (ms *MailSrv) ReadMail(ctx context.Context, req *pbgame.ReadMailRequest,
	rsp *pbgame.DefaultReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	err = mail.ReadMail(u.UserID, req.LogID)
	if err != nil {
		return err
	}
	reply := &pbgame.DefaultReply{
		Result: enumgame.Success,
		LogID:  req.LogID,
	}
	*rsp = *reply
	return nil
}

func (ms *MailSrv) GetMailItems(ctx context.Context, req *pbgame.GetMailItemsRequest,
	rsp *pbgame.GetMailItemsReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	itemList, err := mail.GetMailItems(u.UserID, req.LogID)
	if err != nil {
		return err
	}
	reply := &pbgame.GetMailItemsReply{
		LogID: req.LogID,
	}
	//ItemList: itemList,
	for _, im := range itemList {
		reply.ItemList = append(reply.ItemList, im.ToProto())
	}
	//utilproto.ProtoSlice(itemList, reply.ItemList)
	*rsp = *reply
	return nil
}

func (ms *MailSrv) GetAllMailItems(ctx context.Context, req *pbgame.GetMailItemsRequest,
	rsp *pbgame.GetMailItemsReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	itemList, err := mail.GetAllMailItems(u.UserID)
	if err != nil {
		return err
	}
	reply := &pbgame.GetMailItemsReply{
		Result: 1,
	}
	//utilproto.ProtoSlice(itemList, reply.ItemList)
	for _, im := range itemList {
		reply.ItemList = append(reply.ItemList, im.ToProto())
	}
	*rsp = *reply
	//fmt.Printf("GetAllMailItems:%v\n",itemList)
	return nil
}

func (ms *MailSrv) PagePlayerMail(ctx context.Context, req *pbgame.PagePlayerMailRequest,
	rsp *pbgame.PagePlayerMailReply) error {
	u, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}
	reply, err := mail.PagePlayerMail(req.Page, u.UserID)
	if err != nil {
		return err
	}
	*rsp = *reply
	return nil
}

func (ms *MailSrv) PageMailInfo(ctx context.Context,
	req *pbgame.PageMailInfoRequest, rsp *pbgame.PageMailListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := mail.PageMailInfo(page,
		mdgame.MailInfoFromProto(req.MailInfo))
	if err != nil {
		return err
	}
	for _, mi := range l {
		rsp.List = append(rsp.List, mi.ToProto())
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (ms *MailSrv) PageMailSendLog(ctx context.Context,
	req *pbgame.PageMailSendLogRequest, rsp *pbgame.PageMailSendLogReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := mail.PageMailSendLog(page,
		mdgame.SendMailLogFromProto(req.MailSendLog))
	if err != nil {
		return err
	}
	for _, msl := range l {
		rsp.List = append(rsp.List, msl.ToProto())
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}

func (ms *MailSrv) PageAllPlayerMail(ctx context.Context,
	req *pbgame.PageAllPlayerMailRequest, rsp *pbgame.PageAllPlayerMailReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := mail.PageAllPlayerMail(page,
		mdgame.PlayerMailFromProto(req.PlayerMail))
	if err != nil {
		return err
	}
	for _, pm := range l {
		rsp.List = append(rsp.List, pm.ToProto())
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}
