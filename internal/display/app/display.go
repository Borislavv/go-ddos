package display

import (
	"context"
	displayservice "ddos/internal/display/domain/service"
	"sync"
)

type Display struct {
	ctx      context.Context
	renderer *displayservice.Renderer
}

func New(
	ctx context.Context,
	renderer *displayservice.Renderer,
) *Display {
	return &Display{
		ctx:      ctx,
		renderer: renderer,
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
