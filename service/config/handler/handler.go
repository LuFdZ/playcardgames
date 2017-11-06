package handler

import (
	"playcards/model/config"
	mdconf "playcards/model/config/mod"
	pbconf "playcards/proto/config"
	mdpage "playcards/model/page"
	utilproto "playcards/utils/proto"
	gctx "playcards/utils/context"
	"golang.org/x/net/context"
	utilpb "playcards/utils/proto"
)

type ConfigSrv struct {

}

func NewHandler() *ConfigSrv {
	cs := &ConfigSrv{}
	cs.Init()
	return cs
}
func (cs *ConfigSrv) Init() {
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

func (cs *ConfigSrv) GetConfigs(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.GetConfigsReply) error {
	u := gctx.GetUser(ctx)
	cos := config.GetUniqueConfigByItemID(u.Channel,u.Version,u.MobileOs)
	reply := &pbconf.GetConfigsReply{}
	utilproto.ProtoSlice(reply.List, &cos)
	*rsp = *reply
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
