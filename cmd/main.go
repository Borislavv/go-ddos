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
	"github.com/alexflint/go-arg"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := &config.Config{}
	arg.MustParse(cfg)

	exitCh := make(chan os.Signal, 1)
	defer close(exitCh)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	dur, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer wg.Wait()

	dtCh := make(chan *displaymodel.Table, cfg.MaxRPS)
	smCh := make(chan *displaymodel.Table)

	cl := statservice.NewCollector(cfg)
	st := stat.New(ctx, cfg, dtCh, smCh, cl)
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
