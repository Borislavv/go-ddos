package ddos

import (
	"context"
	"ddos/config"
	service "ddos/internal/ddos/domain/service"
	display "ddos/internal/display/app"
	statservice "ddos/internal/stat/domain/service"
	"github.com/caarlos0/env/v9"
	"log"
	"sync"
	"time"
)

type DDOS struct {
	ctx       context.Context
	cfg       *config.Config
	display   *display.Display
	collector *statservice.Collector
}

func New(ctx context.Context, display *display.Display, collector *statservice.Collector) *DDOS {
	return &DDOS{
		ctx:       ctx,
		cfg:       &config.Config{},
		display:   display,
		collector: collector,
	}
}

func (app *DDOS) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	if err := app.initConfig(); err != nil {
		log.Fatalln(err)
	}

	ctx, cancel, err := app.initCtx()
	if err != nil {
		log.Fatalln(err)
	}
	defer cancel()

	f := service.NewFlooder(ctx, app.cfg, app.display, app.collector)

	f.Run()
}

func (app *DDOS) initConfig() (err error) {
	if err = env.Parse(app.cfg); err != nil {
		return err
	}
	return nil
}

func (app *DDOS) initCtx() (context.Context, context.CancelFunc, error) {
	duration, err := time.ParseDuration(app.cfg.Duration)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(app.ctx, duration)

	return ctx, cancel, nil
}
