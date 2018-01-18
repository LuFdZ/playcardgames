package handler

import (
	"playcards/model/config"
	mdconf "playcards/model/config/mod"
	mdpage "playcards/model/page"
	pbconf "playcards/proto/config"
	gctx "playcards/utils/context"
	utilpb "playcards/utils/proto"
	utilproto "playcards/utils/proto"

	"golang.org/x/net/context"
	"playcards/utils/log"
)

type ConfigSrv struct {
}

func NewHandler() *ConfigSrv {
	cs := &ConfigSrv{}
	cs.init()
	return cs
}
func (cs *ConfigSrv) init() {
	config.RefreshAllConfigsFromDB()
}

func (cs *ConfigSrv) CreateConfig(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.ConfigReply) error {
	co := mdconf.ConfigFromProto(req)
	err := config.CreateConfig(co)
	if err != nil {
		return err
	}

	*rsp = pbconf.ConfigReply{
		Result: 1,
	}
	return nil
}

func (cs *ConfigSrv) UpdateConfig(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.ConfigReply) error {

	co := mdconf.ConfigFromProto(req)
	err := config.UpdateConfig(co)
	if err != nil {
		return err
	}

	*rsp = pbconf.ConfigReply{
		Result: 1,
	}
	return nil
}

func (cs *ConfigSrv) GetConfigsBeforeLogin(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.GetConfigsReply) error {
	//u := gctx.GetUser(ctx)
	cos := config.GetUniqueConfigByItemID(req.Channel, req.Version, req.MobileOs)
	reply := &pbconf.GetConfigsReply{}
	utilproto.ProtoSlice(cos, &reply.List)
	*rsp = *reply
	log.Debug("AAAGetConfigsBeforeLogin:%s|%s|%s\n%v",req.Channel,req.Version,req.MobileOs,rsp)
	return nil
}

func (cs *ConfigSrv) GetConfigs(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.GetConfigsReply) error {
	u := gctx.GetUser(ctx)
	cos := config.GetUniqueConfigByItemID(u.Channel, u.Version, u.MobileOs)
	reply := &pbconf.GetConfigsReply{
		Result: 1,
	}
	utilproto.ProtoSlice(cos, &reply.List)
	*rsp = *reply
	log.Debug("BBBGetConfigsBeforeLogin:%s|%s|%s\n%v",u.Channel, u.Version,u.MobileOs,rsp)
	return nil
}

func (cs *ConfigSrv) RefreshAllConfigsFromDB(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.ConfigReply) error {
	err := config.RefreshAllConfigsFromDB()
	if err != nil {
		return err
	}
	*rsp = pbconf.ConfigReply{
		Result: 1,
	}
	return nil
}

func (rs *ConfigSrv) PageConfigs(ctx context.Context,
	req *pbconf.PageConfigsRequest, rsp *pbconf.PageConfigListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := config.PageConfigs(page,
		mdconf.ConfigFromProto(req.Config))
	if err != nil {
		return err
	}

	err = utilpb.ProtoSlice(l, &rsp.List)
	if err != nil {
		return err
	}
	rsp.Count = rows
	rsp.Result = 1
	return nil
}
