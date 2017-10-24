package sync

import (
	"sync"
	"time"

	"playcards/utils/log"

	gsync "github.com/micro/go-os/sync"
	"github.com/micro/go-plugins/sync/consul"
)

var global gsync.Sync

type GlobalTimer struct {
	stop chan bool
	wg   sync.WaitGroup
}

func NewGlobalTimer() *GlobalTimer {
	return &GlobalTimer{
		stop: make(chan bool),
		wg:   sync.WaitGroup{},
	}
}

func (gt *GlobalTimer) Register(lock string, timeout time.Duration,
	f func() error) {
	go func() {
		gt.wg.Add(1)
		defer gt.wg.Done()

		for {
			select {
			case <-gt.stop:
				log.Info("%s loop stopped", lock)
				return
			case <-time.After(timeout):
				err := GlobalTransaction(lock, f)
				if err != nil {
					log.Err("%s failed: %v", lock, err)
				}
			}
		}
	}()
}

func (gt *GlobalTimer) Stop() {
	close(gt.stop)
	gt.wg.Wait()
	log.Info("global timer stopped")
}

func Init() {
	global = consul.NewSync()
}

func GlobalTransaction(lock string, f func() error) error {
	log.Info("global transcation: %v start", lock)
	l, err := global.Lock(lock)
	if err != nil {
		return err
	}

	if err := l.Acquire(); err != nil {
		return err
	}

	defer func() {
		l.Release()
		log.Info("global transcation: %v unlocked", lock)
	}()

	log.Info("global transcation: %v locked", lock)
	return f()
}

type Once struct {
	m map[string]bool
	l *sync.Mutex
}

func NewOnce() *Once {
	return &Once{
		m: make(map[string]bool),
		l: new(sync.Mutex),
	}
}

func (once *Once) Do(k string, f func() error) error {
	once.l.Lock()
	defer once.l.Unlock()

	_, ok := once.m[k]
	if ok {
		return nil
	}

	once.m[k] = true
	return f()
}
