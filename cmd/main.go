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
	"io"
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

func handleOutput(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, renderer displayservice.Renderer) []io.Writer {
	var writers []io.Writer
	writers = append(writers, renderer)

	if cfg.LogFile != "" {
		f, err := os.Create(cfg.LogFile)
		if err != nil {
			panic(err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				_ = f.Close()
			}
		}()

		writers = append(writers, f)
	}

	return writers
}

func main() {
	cfg, poolCfg := initCfg()

	exitCh := make(chan os.Signal, 1)
	defer close(exitCh)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DurationValue)

	lg := logservice.NewAsync(ctx, cfg)
	lwg := &sync.WaitGroup{}
	lwg.Add(1)
	defer lwg.Wait()
	go lg.Run(lwg)
	defer func() { _ = lg.Close() }()

	cr := func() *http.Client { return &http.Client{Timeout: time.Minute} }
	pl := httpclient.NewPool(ctx, poolCfg, cr)
	defer func() { _ = pl.Close() }()

	cl := statservice.NewCollectorService(ctx, lg, pl, cfg.DurationValue, cfg.Stages)
	rr := displayservice.NewRendererService(ctx, cfg, exitCh, cl)
	sr := sender.NewHttp(cfg, lg, pl, cl)
	mg := workers.NewManagerService(ctx, sr, lg, cl)
	bl := workers.NewBalancerService(ctx, cfg, lg, cl)
	fl := ddosservice.New(ctx, cfg, lg, bl, cl, mg)

	log.SetOutput(logservice.NewMultiWriter(handleOutput(ctx, lwg, cfg, rr)...))

	wg := &sync.WaitGroup{}
	wg.Add(3)
	defer wg.Wait()
	go rr.Run(wg)
	go cl.Run(wg)
	go fl.Run(wg)

	select {
	case <-exitCh:
	case <-ctx.Done():
	}
	lg.Println("interrupting...")
	cancel()
}
