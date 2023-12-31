package logservice

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type CloseFunc func()

type Logger struct {
	ctx    context.Context
	logsCh chan string
}

func NewLogger(ctx context.Context, buffer int) (*Logger, CloseFunc) {
	l := &Logger{
		ctx:    ctx,
		logsCh: make(chan string, buffer),
	}
	return l, l.cls
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

func (l *Logger) Printf(msg string, args ...any) {
	l.logsCh <- fmt.Sprintf(msg, args...)
}

func (l *Logger) cls() {
	close(l.logsCh)
}
