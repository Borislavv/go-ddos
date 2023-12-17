package main

import (
	"context"
	"ddos/internal/ddos/app"
	display "ddos/internal/display/app"
	displayservice "ddos/internal/display/domain/service"
	stat "ddos/internal/stat/app"
	statservice "ddos/internal/stat/domain/service"
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

	cl := statservice.NewCollector()
	st := stat.New(ctx)
	di := display.New(ctx, displayservice.NewRenderer(ctx, cl, exitCh), exitCh)
	dd := ddos.New(ctx, di, cl)

	wg.Add(3)
	go st.Run(wg)
	go di.Run(wg)
	go dd.Run(wg)

	<-exitCh
	cancel()
	wg.Wait()
}
