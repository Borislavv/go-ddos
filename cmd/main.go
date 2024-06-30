package main

import (
	"context"
	"github.com/Borislavv/go-ddos/config"
	displayservice "github.com/Borislavv/go-ddos/internal/display/domain/service"
	ddosservice "github.com/Borislavv/go-ddos/internal/flooder/app"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/orchestrator"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/worker"
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
	cfg.Validate()
	return cfg, cfg.HttpClinePoolConfig()
}

func logWriters(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, renderer displayservice.Renderer) []io.Writer {
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

	cl := statservice.NewCollectorService(lg, pl, cfg.DurationValue, cfg.Stages)
	rr := displayservice.NewRendererService(cfg, exitCh, cl)
	sr := sender.NewHttp(cfg, lg, pl, cl)
	mg := worker.NewManagerService(ctx, sr, lg, cl)
	bl := worker.NewBalancerService(ctx, cfg, lg, cl)
	or := orchestrator.NewWorkersOrchestrator(cfg, mg, bl, lg)
	fl := ddosservice.New(cfg, or, lg)

	log.SetOutput(logservice.NewMultiWriter(logWriters(ctx, lwg, cfg, rr)...))

	wg := &sync.WaitGroup{}
	wg.Add(3)
	defer wg.Wait()
	go rr.Run(ctx, wg)
	go cl.Run(ctx, wg)
	go fl.Run(ctx, wg)

	select {
	case <-exitCh:
	case <-ctx.Done():
	}
	lg.Println("interrupting...")
	cancel()
}
