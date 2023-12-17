package display

import (
	"context"
	displayservice "ddos/internal/display/domain/service"
	"os"
	"sync"
)

type Display struct {
	ctx      context.Context
	renderer *displayservice.Renderer
	exitCh   chan os.Signal
}

func New(
	ctx context.Context,
	renderer *displayservice.Renderer,
	exitCh chan os.Signal,
) *Display {
	return &Display{
		ctx:      ctx,
		renderer: renderer,
		exitCh:   exitCh,
	}
}

func (d *Display) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go d.renderer.Draw(wg)
	go d.renderer.Close(wg)
	wg.Wait()
}
