package logservice

import (
	"context"
	"ddos/config"
	"log"
	"sync"
)

type Logger struct {
	ctx    context.Context
	cfg    *config.Config
	logsCh chan string
}

func NewLogger(ctx context.Context, cfg *config.Config, logsCh chan string) *Logger {
	return &Logger{
		ctx:    ctx,
		cfg:    cfg,
		logsCh: logsCh,
	}
}

func (l *Logger) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	for {
		select {
		case <-l.ctx.Done():
			return
		case msg := <-l.logsCh:
			log.Println(msg)
		}
	}
}

func (l *Logger) Println(msg string) {
	l.logsCh <- msg
}
