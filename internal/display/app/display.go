package display

import (
	"context"
	displayservice "github.com/Borislavv/go-ddos/internal/display/domain/service"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"sync"
)

type App struct {
	ctx      context.Context
	logger   logservice.Logger
	renderer displayservice.Renderer
}

func New(
	ctx context.Context,
	logger logservice.Logger,
	renderer displayservice.Renderer,
) *App {
	return &App{
		ctx:      ctx,
		logger:   logger,
		renderer: renderer,
	}
}

func (d *App) Run(mwg *sync.WaitGroup) {
	defer func() {
		d.logger.Println("display.App.Run() is closed")
		mwg.Done()
	}()

	ctx, cancel := context.WithCancel(d.ctx)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go d.renderer.Draw(wg, ctx)
	go d.renderer.Listen(wg, cancel)
	wg.Wait()
}
