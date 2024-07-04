package flooder

import (
	"context"
	ddos "github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/orchestrator"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"sync"
)

type App struct {
	cfg          *ddos.Config
	orchestrator orchestrator.Orchestrator
	logger       logservice.Logger
}

func New(
	cfg *ddos.Config,
	orchestrator orchestrator.Orchestrator,
	logger logservice.Logger,
) *App {
	return &App{
		cfg:          cfg,
		logger:       logger,
		orchestrator: orchestrator,
	}
}

func (f *App) Run(ctx context.Context, mwg *sync.WaitGroup) {
	defer mwg.Done()
	defer f.logger.Println("flooder.App.Run(): is closed")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go f.orchestrator.Run(ctx, wg)
}
