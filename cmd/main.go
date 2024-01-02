package main

import (
	"context"
	"ddos/config"
	display "ddos/internal/display/app"
	displayservice "ddos/internal/display/domain/service"
	ddosservice "ddos/internal/flooder/app"
	"ddos/internal/flooder/domain/service/sender"
	"ddos/internal/flooder/domain/service/workers"
	"ddos/internal/flooder/infrastructure/httpclient"
	httpclientconfig "ddos/internal/flooder/infrastructure/httpclient/config"
	logservice "ddos/internal/log/domain/service"
	stat "ddos/internal/stat/app"
	statservice "ddos/internal/stat/domain/service"
	"github.com/alexflint/go-arg"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func initCfg() (*config.Config, *httpclientconfig.Config) {
	cfg := &config.Config{}
	arg.MustParse(cfg)

	poolCfg := &httpclientconfig.Config{
		PoolInitSize: cfg.PoolInitSize,
		PoolMaxSize:  cfg.PoolMaxSize,
	}

	return cfg, poolCfg
}

func main() {
	cfg, poolCfg := initCfg()

	exitCh := make(chan os.Signal, 1)
	defer close(exitCh)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	duration, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)

	lg := logservice.NewAsync(ctx, cfg)
	lwg := &sync.WaitGroup{}
	lwg.Add(1)
	defer lwg.Wait()
	go lg.Run(lwg)
	defer func() { _ = lg.Close() }()

	cr := func() *http.Client { return &http.Client{Timeout: time.Minute} }
	pl := httpclient.NewPool(ctx, poolCfg, cr)
	defer func() { _ = pl.Close() }()

	cl := statservice.NewCollectorService(ctx, lg, pl, duration, cfg.Stages)
	rr := displayservice.NewRendererService(ctx, lg, exitCh)
	st := stat.New(ctx, cfg, lg, rr, cl)
	dy := display.New(ctx, lg, rr)
	sr := sender.NewHttp(cfg, lg, pl, cl)
	mg := workers.NewManagerService(ctx, sr, lg, cl)
	bl := workers.NewBalancerService(ctx, cfg, cl)
	fl := ddosservice.New(ctx, cfg, lg, bl, cl, mg)

	wg := &sync.WaitGroup{}
	wg.Add(4)
	defer wg.Wait()
	go cl.Run(wg)
	go dy.Run(wg)
	go st.Run(wg)
	go fl.Run(wg)

	<-exitCh
	cancel()
}
