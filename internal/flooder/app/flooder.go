package flooder

import (
	"context"
	ddos "github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"math"
	"sync"
	"time"
)

type App struct {
	ctx       context.Context
	mu        *sync.RWMutex
	cfg       *ddos.Config
	manager   workers.Manager
	balancer  workers.Balancer
	logger    logservice.Logger
	collector statservice.Collector
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
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		cfg:       cfg,
		logger:    logger,
		manager:   manager,
		collector: collector,
		balancer:  reqBalancer,
	}
}

func (f *App) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()
	defer f.logger.Println("flooder.App.Run(): is closed")

	balancerTicker := time.NewTicker(time.Millisecond * 100)
	defer balancerTicker.Stop()

	reqSendTicker := time.NewTicker(time.Duration(math.Ceil(float64(time.Second) / float64(f.cfg.MaxRPS))))
	defer reqSendTicker.Stop()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(f.ctx)

	for {
		select {
		case <-f.ctx.Done():
			f.manager.CloseAll(cancel, wg)
			return
		case <-balancerTicker.C:
			action, sleep := f.balancer.CurrentAction()
			switch action {
			case enum.Spawn:
				f.manager.SpawnOne(ctx, wg, reqSendTicker)
			case enum.Close:
				f.manager.CloseOne()
			case enum.Await:
				time.Sleep(sleep)
			}
		}
	}
}
