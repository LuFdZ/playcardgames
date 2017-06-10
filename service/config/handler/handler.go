package handler

import (
	"playcards/model/config"
	mdconf "playcards/model/config/mod"
	pbconf "playcards/proto/config"
	utilproto "playcards/utils/proto"

	"golang.org/x/net/context"
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
