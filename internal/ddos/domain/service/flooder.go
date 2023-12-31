package ddosservice

import (
	"context"
	ddos "ddos/config"
	reqsender "ddos/internal/ddos/domain/service/balancer/req"
	"ddos/internal/ddos/domain/service/manager/req"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"runtime"
	"sync"
	"time"
)

type Flooder struct {
	mu  *sync.RWMutex
	ctx context.Context

	cfg         *ddos.Config
	manager     *req.Manager
	logger      *logservice.Logger
	collector   *statservice.Collector
	reqBalancer *reqsender.Balancer
}

func NewFlooder(
	ctx context.Context,
	cfg *ddos.Config,
	manager *req.Manager,
	logger *logservice.Logger,
	reqBalancer *reqsender.Balancer,
	collector *statservice.Collector,
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
	defer mwg.Done()

	wg := &sync.WaitGroup{}
	defer wg.Wait()

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
