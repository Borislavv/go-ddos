package logservice

import (
	"context"
	"ddos/config"
	"fmt"
	"log"
	"os"
	"sync"
)

type CloseFunc func()

type Async struct {
	ctx context.Context
	ch  chan string
	fd  *os.File
}

func NewAsync(ctx context.Context, cfg *config.Config) (*Async, CloseFunc) {
	l := &Async{
		ctx: ctx,
		ch:  make(chan string, cfg.MaxRPS*int(cfg.MaxWorkers)),
	}

	if cfg.LogFile != "" {
		f, err := os.Create(cfg.LogFile)
		if err != nil {
			panic(err)
		}
		l.fd = f
		log.SetOutput(f)
	}

	return l, l.cls
}

func (l *Async) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-l.ctx.Done():
			return
		case msg := <-l.ch:
			log.Println(msg)
		}
	}
}

func (l *Async) Println(msg string) {
	l.ch <- msg
}

func (l *Async) Printf(msg string, args ...any) {
	l.ch <- fmt.Sprintf(msg, args...)
}

func (l *Async) cls() {
	close(l.ch)
	_ = l.fd.Close()
}
