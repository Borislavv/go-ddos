package main

import (
	"context"
	"github.com/Borislavv/go-ddos/config"
	displayservice "github.com/Borislavv/go-ddos/internal/display/domain/service"
	ddosservice "github.com/Borislavv/go-ddos/internal/flooder/app"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient"
	httpclientconfig "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/config"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"github.com/alexflint/go-arg"
	"log"
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

	testDuration, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		panic(err)
	}
	cfg.DurationValue = testDuration

	targetAvgSuccessRequestsDuration, err := time.ParseDuration(cfg.TargetAvgSuccessRequestsDuration)
	if err != nil {
		panic(err)
	}
	cfg.TargetAvgSuccessRequestsDurationValue = targetAvgSuccessRequestsDuration

	reqSenderSpawnInterval, err := time.ParseDuration(cfg.SpawnInterval)
	if err != nil {
		panic(err)
	}
	cfg.SpawnIntervalValue = reqSenderSpawnInterval

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

	r := displayservice.NewRendererV2Service(ctx, cfg, exitCh, cl)
	log.SetOutput(r)

	//rr := displayservice.NewRendererService(ctx, lg, exitCh)
	//st := stat.New(ctx, cfg, lg, rr, cl)
	//dy := display.New(ctx, lg, rr)
	sr := sender.NewHttp(cfg, lg, pl, cl)
	mg := workers.NewManagerService(ctx, sr, lg, cl)
	bl := workers.NewBalancerService(ctx, cfg, lg, cl)
	fl := ddosservice.New(ctx, cfg, lg, bl, cl, mg)

	wg := &sync.WaitGroup{}
	wg.Add(3)
	defer wg.Wait()
	go r.Run(wg)
	time.Sleep(time.Second)
	go cl.Run(wg)
	//go dy.Run(wg)
	//go st.Run(wg)
	go fl.Run(wg)

	select {
	case <-exitCh:
	case <-ctx.Done():
	}
	lg.Println("interrupting...")
	cancel()
}
