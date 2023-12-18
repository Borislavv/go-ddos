package ddos

import (
	"context"
	"ddos/config"
	service "ddos/internal/ddos/domain/service"
	display "ddos/internal/display/app"
	statservice "ddos/internal/stat/domain/service"
	"sync"
)

type DDOS struct {
	ctx       context.Context
	cfg       *config.Config
	display   *display.Display
	collector *statservice.Collector
}

func New(ctx context.Context, cfg *config.Config, display *display.Display, collector *statservice.Collector) *DDOS {
	return &DDOS{
		ctx:       ctx,
		cfg:       cfg,
		display:   display,
		collector: collector,
	}
}

func (app *DDOS) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	f := service.NewFlooder(app.ctx, app.cfg, app.display, app.collector)

	f.Run()
}
