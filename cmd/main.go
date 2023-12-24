package main

import (
	"context"
	"ddos/config"
	"ddos/internal/ddos/app"
	display "ddos/internal/display/app"
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

	if cfg.StdOutFile != "" {
		outfile, err := os.Create(cfg.StdOutFile)
		if err != nil {
			panic(err)
		}
		defer func() { _ = outfile.Close() }()
		os.Stdout = outfile
	}

	if cfg.StdErrFile != "" {
		errfile, err := os.Create(cfg.StdErrFile)
		if err != nil {
			panic(err)
		}
		defer func() { _ = errfile.Close() }()
		os.Stderr = errfile
	}

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

	cl := statservice.NewCollector(cfg)
	rd := displayservice.NewRenderer(ctx, cfg, exitCh)
	st := stat.New(ctx, cfg, rd, cl)
	di := display.New(ctx, rd)
	dd := ddos.New(ctx, cfg, di, cl)

	wg.Add(3)
	go st.Run(wg)
	go di.Run(wg)
	go dd.Run(wg)

	<-exitCh
	cancel()
}
