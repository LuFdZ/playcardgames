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

func (cs *ConfigSrv) UpdateConfig(ctx context.Context,
	req *pbconf.Config, rsp *pbconf.Config) error {

	conf := mdconf.ConfigFromProto(req)
	err := config.UpdateConfig(conf)
	if err != nil {
		return err
	}

	*rsp = *conf.ToProto()

	return nil
}

func (cs *ConfigSrv) ConfigList(ctx context.Context,
	req *pbconf.ConfigListRequest, rsp *pbconf.ConfigListReply) error {

	rs, err := config.ConfigList()
	if err != nil {
		return err
	}

	err = utilproto.ProtoSlice(rs, &rsp.List)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ConfigSrv) GetConfigByID(ctx context.Context,
	req *pbconf.GetConfigByIDRequest, rsp *pbconf.Config) error {
	rs, err := config.GetConfigByID(req.ID)
	if err != nil {
		return err
	}
	*rsp = *rs.ToProto()
	return nil
}

func (cs *ConfigSrv) CreateConfigOpen(ctx context.Context,
	req *pbconf.ConfigOpen, rsp *pbconf.ConfigReply) error {

	co := mdconf.ConfigOpenFromProto(req)
	err := config.CreateConfigOpen(co)
	if err != nil {
		return err
	}

	rsp = &pbconf.ConfigReply{
		Result: co.OpenID,
	}
	return nil
}

func (cs *ConfigSrv) UpdateConfigOpen(ctx context.Context,
	req *pbconf.ConfigOpen, rsp *pbconf.ConfigReply) error {

	co := mdconf.ConfigOpenFromProto(req)
	err := config.UpdateConfigOpen(co)
	if err != nil {
		return err
	}

	rsp = &pbconf.ConfigReply{
		Result: 1,
	}
	return nil
}

func (cs *ConfigSrv) GetConfigOpens(ctx context.Context,
	req *pbconf.GetConfigOpensRequest, rsp *pbconf.GetConfigOpensReply) error {
	u := gctx.GetUser(ctx)
	cos := config.GetConfigOpens(u.Channel,req.Version,u.MobileOs)
	reply := &pbconf.GetConfigOpensReply{}
	utilproto.ProtoSlice(reply.List, &cos)
	*rsp = *reply
	return nil
}

func (cs *ConfigSrv) RefreshAllConfigOpensFromDB(ctx context.Context,
	req *pbconf.ConfigOpen, rsp *pbconf.ConfigReply) error {
	err := config.RefreshAllConfigOpensFromDB()
	if err != nil {
		return err
	}
	rsp = &pbconf.ConfigReply{
		Result: 1,
	}
	return nil
}

func (rs *ConfigSrv) PageConfigOpens(ctx context.Context,
	req *pbconf.PageConfigOpensRequest, rsp *pbconf.PageNoticeListReply) error {
	page := mdpage.PageOptionFromProto(req.Page)
	rsp.Result = 2
	l, rows, err := config.PageConfigOpens(page,
		mdconf.ConfigOpenFromProto(req.ConfigOpen))
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