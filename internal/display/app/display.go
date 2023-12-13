package display

import (
	"context"
	model "ddos/internal/display/domain/model"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
)

type Display struct {
	ctx    context.Context
	dataCh chan *model.Table
}

func New(ctx context.Context) *Display {
	return &Display{ctx: ctx, dataCh: make(chan *model.Table, 1000)}
}

func (d *Display) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go d.draw(wg)
	go d.close(wg)
	wg.Wait()
}

func (d *Display) Draw(t *model.Table) {
	d.dataCh <- t
}

func (d *Display) close(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyCtrlC || ev.Key == termbox.KeyCtrlZ {
					termbox.Close()
					return
				}
			case termbox.EventError:
				panic(ev.Err)
			}
		}
	}
}

func (d *Display) draw(wg *sync.WaitGroup) {
	defer wg.Done()

	table := tablewriter.NewWriter(os.Stdout)

	for {
		select {
		case <-d.ctx.Done():
			termbox.Close()
			return
		case data := <-d.dataCh:
			if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				log.Fatalln(err)
			}
			if err := termbox.Sync(); err != nil {
				log.Fatalln(err)
			}

			table.SetHeader(data.Header)
			table.ClearRows()

			for _, row := range data.Rows {
				table.Append(row)
			}

			table.Render()
			if err := termbox.Flush(); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
