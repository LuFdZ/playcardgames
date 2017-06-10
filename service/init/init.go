package envinit

import (
	"playcards/utils/cache"
	gcf "playcards/utils/config"
	"playcards/utils/db"
	"playcards/utils/env"
	glog "playcards/utils/log"
	"playcards/utils/sync"

	"github.com/google/gops/agent"
	"github.com/micro/go-os/config"
	"github.com/micro/go-os/config/source/consul"
	_ "github.com/micro/go-plugins/broker/nats"

	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

var Debug = false

func Init() {
	cf := config.NewConfig(config.WithSource(
		consul.NewSource(config.SourceName("/bcr/config"))),
	)

	gcf.Init(cf)

	Debug = gcf.Debug()
	logLevel := gcf.LogLevel()
	logger, err := glog.NewLogger(env.LogPath, logLevel)
	env.ErrExit(err)

	glog.SetDefault(logger)

	dburl := gcf.DBURL()
	err = db.Open(Debug, "mysql", dburl)
	env.ErrExit(err)

	redishost := gcf.RedisHost()[0]
	cache.Init(redishost)
	sync.Init()

	InitPProfServices()
}

func InitPProfServices() {
	if Debug {
		runtime.SetBlockProfileRate(1)
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Println("start pprof service failed")
		return
	}

	log.Println("start pprof service on:", ln.Addr())
	go func() {
		http.Serve(ln, nil)
	}()

	opts := &agent.Options{}
	if err := agent.Listen(opts); err != nil {
		log.Println("start gops agent failed", err)
	}
}
