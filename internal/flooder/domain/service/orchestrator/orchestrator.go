package orchestrator

import (
	"context"
	ddos "github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/worker"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"math"
	"time"
)

type WorkersOrchestrator struct {
	cfg      *ddos.Config
	manager  worker.Manager
	balancer worker.Balancer
	logger   logservice.Logger
}

func NewWorkersOrchestrator(
	cfg *ddos.Config,
	manager worker.Manager,
	balancer worker.Balancer,
	logger logservice.Logger,
) *WorkersOrchestrator {
	return &WorkersOrchestrator{
		cfg:      cfg,
		manager:  manager,
		balancer: balancer,
		logger:   logger,
	}
}

func (o *WorkersOrchestrator) Run(ctx context.Context) {
	defer o.logger.Println("flooder.WorkersOrchestrator.Run(): is closed")

	balancerTicker := time.NewTicker(time.Millisecond * 100)
	defer balancerTicker.Stop()

	reqSendTicker := time.NewTicker(time.Duration(math.Ceil(float64(time.Second) / float64(o.cfg.MaxRPS))))
	defer reqSendTicker.Stop()

	ctx, cancel := context.WithCancel(ctx)

	for {
		select {
		case <-ctx.Done():
			o.manager.CloseAll(cancel)
			return
		case <-balancerTicker.C:
			action, sleep := o.balancer.CurrentAction()
			switch action {
			case enum.Spawn:
				o.manager.SpawnOne(ctx, reqSendTicker)
			case enum.Close:
				o.manager.CloseOne()
			case enum.Await:
				time.Sleep(sleep)
			}
		}
	}
}
