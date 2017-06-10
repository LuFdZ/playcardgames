package subscribe

import (
	"playcards/service/web/clients"
	"playcards/utils/log"
	"playcards/utils/sync"

	"github.com/micro/go-micro/broker"
)

var once *sync.Once

func init() {
	once = sync.NewOnce()
}

func SrvSubscribe(b broker.Broker, t string, h broker.Handler) error {
	return once.Do(t, func() error {
		log.Info("service subscribe %v", t)
		clients.InitTopic(t)
		_, err := b.Subscribe(t, h)
		return err
	})
}
