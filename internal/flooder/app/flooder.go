package flooder

import (
	"context"
	ddos "github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"sync"
	"time"
)

type App struct {
	ctx         context.Context
	mu          *sync.RWMutex
	cfg         *ddos.Config
	manager     workers.Manager
	reqBalancer workers.Balancer
	logger      logservice.Logger
	collector   statservice.Collector
}

func New(
	ctx context.Context,
	cfg *ddos.Config,
	logger logservice.Logger,
	reqBalancer workers.Balancer,
	collector statservice.Collector,
	manager workers.Manager,
) *App {
	return &App{
		mu:          &sync.RWMutex{},
		ctx:         ctx,
		cfg:         cfg,
		logger:      logger,
		manager:     manager,
		collector:   collector,
		reqBalancer: reqBalancer,
	}
}

func (f *App) Run(mwg *sync.WaitGroup) {
	defer func() {
		f.logger.Println("flooder.App.Run(): is closed")
		mwg.Done()
	}()

	balancerTicker := time.NewTicker(time.Millisecond * 100)
	defer balancerTicker.Stop()

	reqSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.cfg.MaxRPS)*(1+f.cfg.ToleranceCoefficient)))
	defer reqSendTicker.Stop()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(f.ctx)

	for {
		select {
		case <-f.ctx.Done():
			f.manager.CloseAll(cancel, wg)
			return
		case <-balancerTicker.C:
			if f.reqBalancer.IsMustBeSpawned() {
				f.manager.Spawn(ctx, wg, reqSendTicker)
			} else if f.reqBalancer.IsMustBeClosed() {
				f.manager.CloseOne()
			}
		}
	}
}
