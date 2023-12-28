package ddos

import (
	"context"
	"ddos/config"
	service "ddos/internal/ddos/domain/service"
	display "ddos/internal/display/app"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"sync"
)

type DDOS struct {
	ctx       context.Context
	cfg       *config.Config
	display   *display.Display
	logger    *logservice.Logger
	collector *statservice.Collector
}

func New(
	ctx context.Context,
	cfg *config.Config,
	display *display.Display,
	logger *logservice.Logger,
	collector *statservice.Collector,
) *DDOS {
	return &DDOS{
		ctx:       ctx,
		cfg:       cfg,
		display:   display,
		logger:    logger,
		collector: collector,
	}
}

func (app *DDOS) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	service.
		NewFlooder(app.ctx, app.cfg, app.logger, app.collector).
		Run()
}
