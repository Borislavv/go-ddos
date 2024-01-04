package logservice

import (
	"context"
	"ddos/config"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

type CloseFunc func()

type Async struct {
	ctx context.Context
	ch  chan string
	fd  *os.File
}

func NewAsync(ctx context.Context, cfg *config.Config) *Async {
	buff := int64(math.Ceil(float64(cfg.MaxRPS) / float64(cfg.MaxWorkers)))
	if buff <= 0 {
		buff = 1
	}

	l := &Async{
		ctx: ctx,
		ch:  make(chan string, buff),
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

	for msg := range l.ch {
		log.Println(msg)
	}
}

func (l *Async) Println(msg string) {
	l.ch <- msg
}

func (l *Async) Printf(msg string, args ...any) {
	l.Println(fmt.Sprintf(msg, args...))
}

func (l *Async) Close() error {
	l.Println("logger.Async.Run() is closed")
	if len(l.ch) > 0 {
		t := time.NewTimer(time.Second * 3)
		defer t.Stop()
		<-t.C
	}
	close(l.ch)
	if l.fd != nil {
		if err := l.fd.Close(); err != nil {
			return err
		}
	}
	log.SetOutput(os.Stdout)
	return nil
}
