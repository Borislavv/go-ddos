package ddosservice

import (
	"context"
	ddos "ddos/config"
	reqsender "ddos/internal/ddos/domain/service/balancer/req"
	"ddos/internal/ddos/domain/service/manager/ddosmanagerservice"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"runtime"
	"sync"
	"time"
)

type Flooder struct {
	mu          *sync.RWMutex
	ctx         context.Context
	cfg         *ddos.Config
	logger      logservice.Logger
	reqBalancer *reqsender.Balancer
	collector   statservice.Collector
	manager     *ddosmanagerservice.Manager
}

func NewFlooder(
	ctx context.Context,
	cfg *ddos.Config,
	logger logservice.Logger,
	reqBalancer *reqsender.Balancer,
	collector statservice.Collector,
	manager *ddosmanagerservice.Manager,
) *Flooder {
	return &Flooder{
		mu:          &sync.RWMutex{},
		ctx:         ctx,
		cfg:         cfg,
		logger:      logger,
		manager:     manager,
		collector:   collector,
		reqBalancer: reqBalancer,
	}
}

func (f *Flooder) Run(mwg *sync.WaitGroup) {
	defer func() {
		f.logger.Println("ddos.Flooder.Run() is closed")
		mwg.Done()
	}()

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
		f.logger.Println("ddos.Flooder.Workers all spawned are closed")
	}()

	balancerTicker := time.NewTicker(time.Millisecond * 100)
	defer balancerTicker.Stop()

	reqSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.cfg.MaxRPS)*1.11))
	defer reqSendTicker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			f.manager.CloseAll()
			return
		case <-balancerTicker.C:
			if f.reqBalancer.IsMustBeSpawned() {
				f.manager.Spawn(reqSendTicker, wg)
			} else if f.reqBalancer.IsMustBeClosed() {
				f.manager.Close()
			} else {
				runtime.Gosched()
			}
		}
	}
}
