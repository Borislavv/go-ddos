package main

import (
	"context"
	"ddos/internal/ddos/app"
	app2 "ddos/internal/display/app"
	"log"
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

	dataCh := make(chan *app2.Data, 1000)

	wg.Add(2)
	go app.NewDDOS(ctx, dataCh).Run(wg)
	go app2.NewDisplay(ctx, dataCh).Run(wg)

	log.Println("awaiting ctrl+c")
	<-exit
	log.Println("received ctrl+c sing, closing...")
	cancel()
	wg.Wait()
	log.Println("exited!")
}
