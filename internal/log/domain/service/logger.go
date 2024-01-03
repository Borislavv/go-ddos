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

func NewAsync(ctx context.Context, cfg *config.Config) *Async {
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

	return l
}

func (l *Async) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-l.ctx.Done():
			go func() {
				for range l.ch {
					// clean remain messages into the chan
					// due to non-block other goroutines
				}
			}()
			return
		case msg := <-l.ch:
			log.Println(msg)
		}
	}
}

func (l *Async) Println(msg string) {
	select {
	case l.ch <- msg:
	case <-l.ctx.Done():
	default:
		fmt.Println(msg)
	}
}

func (l *Async) Printf(msg string, args ...any) {
	l.Println(fmt.Sprintf(msg, args...))
}

func (l *Async) Close() error {
	if l.fd != nil {
		if err := l.fd.Close(); err != nil {
			return err
		}
	}
	close(l.ch)
	return nil
}
