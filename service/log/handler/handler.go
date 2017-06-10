package handler

import (
	mlog "playcards/model/log"
	mdlog "playcards/model/log/mod"
	pblog "playcards/proto/log"
	gctx "playcards/utils/context"
	"playcards/utils/log"

	"github.com/micro/go-micro"
	"golang.org/x/net/context"
)

type LogSrv struct {
	cerror chan *mdlog.ClientErrorLog
}

func NewHandler(srv micro.Service) *LogSrv {
	ls := &LogSrv{
		cerror: make(chan *mdlog.ClientErrorLog, 1024),
	}
	go ls.writeLogLoop()
	return ls
}

func (ls *LogSrv) writeLogLoop() {
	for {
		select {
		case cerrlog := <-ls.cerror:
			err := mlog.AddClientErrorLog(cerrlog)
			if err != nil {
				log.Warn("save client error log error : %v", err)
			}
		}
	}
}

func (ls *LogSrv) ClientReportException(ctx context.Context,
	req *pblog.ClientErrorLog, rsp *pblog.ClientErrorLog) error {
	u := gctx.GetUser(ctx)
	address, _ := ctx.Value("X-Real-Ip").(string)

	errlog := mdlog.ClientErrorLogFromProto(req)
	errlog.ClientAddress = address
	if u != nil {
		errlog.UserID = u.UserID
	}

	ls.cerror <- errlog
	return nil
}
