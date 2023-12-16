package display

import (
	"context"
	model "ddos/internal/display/domain/model"
	displayservice "ddos/internal/display/domain/service"
	"os"
	"sync"
)

type Display struct {
	ctx    context.Context
	dataCh chan *model.Table
	exitCh chan os.Signal
}

func New(ctx context.Context, exitCh chan os.Signal) *Display {
	return &Display{
		ctx:    ctx,
		exitCh: exitCh,
		dataCh: make(chan *model.Table, 1000),
	}
}

func (d *Display) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	renderer := displayservice.NewRenderer(d.ctx, d.exitCh, d.dataCh)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go renderer.Draw(wg)
	go renderer.Close(wg)
	wg.Wait()
}

func (d *Display) Draw(t *model.Table) {
	d.dataCh <- t
}
