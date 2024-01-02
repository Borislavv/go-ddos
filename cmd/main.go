package main

import (
	"context"
	"ddos/config"
	ddosservice "ddos/internal/ddos/domain/service"
	reqsender "ddos/internal/ddos/domain/service/balancer/req"
	"ddos/internal/ddos/domain/service/manager/req"
	"ddos/internal/ddos/domain/service/sender"
	"ddos/internal/ddos/infrastructure/httpclient"
	httpclientconfig "ddos/internal/ddos/infrastructure/httpclient/config"
	display "ddos/internal/display/app"
	displayservice "ddos/internal/display/domain/service"
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
	mainCfg, poolCfg := initCfg()

	exitCh := make(chan os.Signal, 1)
	defer close(exitCh)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	duration, err := time.ParseDuration(mainCfg.Duration)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)

	lg, loggerCls := logservice.NewAsync(ctx, mainCfg)
	defer loggerCls()

	creator := func() *http.Client { return &http.Client{Timeout: time.Minute} }
	pool, poolCls := httpclient.NewPool(ctx, poolCfg, creator)
	defer poolCls()

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	cl := statservice.NewCollector(ctx, pool, duration, mainCfg.Stages)
	rr := displayservice.NewRenderer(ctx, mainCfg, exitCh)
	st := stat.New(ctx, mainCfg, rr, cl)
	dy := display.New(ctx, rr)
	sr := sender.NewSender(mainCfg, lg, pool, cl)
	mg := req.NewManager(ctx, sr, cl)
	bl := reqsender.NewBalancer(ctx, mainCfg, cl)
	fl := ddosservice.NewFlooder(ctx, mainCfg, mg, lg, bl, cl)

	wg.Add(5)
	go lg.Run(wg)
	go cl.Run(wg)
	go st.Run(wg)
	go dy.Run(wg)
	go fl.Run(wg)

	<-exitCh
	cancel()
}
