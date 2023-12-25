package main

import (
	"context"
	"ddos/config"
	"ddos/internal/ddos/app"
	display "ddos/internal/display/app"
	displayservice "ddos/internal/display/domain/service"
	logservice "ddos/internal/log/domain/service"
	stat "ddos/internal/stat/app"
	statservice "ddos/internal/stat/domain/service"
	"github.com/alexflint/go-arg"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := &config.Config{}
	arg.MustParse(cfg)

	if cfg.LogFile != "" {
		logfile, err := os.Create(cfg.LogFile)
		if err != nil {
			panic(err)
		}
		defer func() { _ = logfile.Close() }()
		log.SetOutput(logfile)
	}

	logsCh := make(chan string, int64(cfg.MaxRPS)*cfg.MaxWorkers)
	exitCh := make(chan os.Signal, 1)
	defer func() { close(exitCh); close(logsCh) }()
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	dur, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer wg.Wait()

	lg := logservice.NewLogger(ctx, cfg, logsCh)
	cl := statservice.NewCollector(cfg)
	rd := displayservice.NewRenderer(ctx, cfg, exitCh)
	st := stat.New(ctx, cfg, rd, cl)
	di := display.New(ctx, rd)
	dd := ddos.New(ctx, cfg, di, lg, cl)

	wg.Add(4)
	go lg.Run(wg)
	go st.Run(wg)
	go di.Run(wg)
	go dd.Run(wg)

	<-exitCh
	cancel()
}
