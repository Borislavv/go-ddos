package main

import (
	"context"
	"ddos/config"
	"ddos/internal/ddos/app"
	display "ddos/internal/display/app"
	displaymodel "ddos/internal/display/domain/model"
	displayservice "ddos/internal/display/domain/service"
	stat "ddos/internal/stat/app"
	statservice "ddos/internal/stat/domain/service"
	"github.com/caarlos0/env/v9"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	exitCh := make(chan os.Signal, 1)
	defer close(exitCh)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer wg.Wait()

	cfg, err := initConfig()
	if err != nil {
		panic(err)
	}

	dtCh := make(chan *displaymodel.Table, cfg.MaxRPS)
	smCh := make(chan *displaymodel.Table)

	cl := statservice.NewCollector()
	st := stat.New(ctx, dtCh, smCh, cl)
	rd := displayservice.NewRenderer(ctx, dtCh, smCh, exitCh)
	di := display.New(ctx, rd)
	dd := ddos.New(ctx, cfg, di, cl)

	wg.Add(3)
	go st.Run(wg)
	go di.Run(wg)
	go dd.Run(wg)

	<-exitCh
	cancel()
}

func initConfig() (cfg *config.Config, err error) {
	cfg = &config.Config{}
	if err = env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
