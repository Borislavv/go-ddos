package main

import (
	"context"
	"ddos/internal/ddos/app"
	display "ddos/internal/display/app"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	exit := make(chan os.Signal, 1)
	defer close(exit)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer wg.Wait()

	di := display.New(ctx, exit)
	dd := ddos.New(ctx, di)

	wg.Add(2)
	go di.Run(wg)
	go dd.Run(wg)

	<-exit
	cancel()
	wg.Wait()
}
