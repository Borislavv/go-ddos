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
	mu     *sync.Mutex
	ctx    context.Context
	exit   chan os.Signal
	dataCh chan *model.Table
	table  *tablewriter.Table
	stopCh chan struct{}
}

func New(ctx context.Context, exit chan os.Signal) *Display {
	return &Display{
		mu:     &sync.Mutex{},
		ctx:    ctx,
		exit:   exit,
		dataCh: make(chan *model.Table, 1000),
		stopCh: make(chan struct{}, 1),
	}
}

func (d *Display) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	d.table = tablewriter.NewWriter(os.Stdout)

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
	defer func() {
		close(d.stopCh)
		wg.Done()
	}()

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyCtrlC || ev.Key == termbox.KeyCtrlZ {
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

	for {
		select {
		case <-d.stopCh:
			termbox.Close()

			d.exit <- os.Interrupt

			// draw the summary table
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"hello"})
			table.Append([]string{"world"})

			// render the summary table
			table.Render()

			// draw the summary table
			if err := termbox.Flush(); err != nil {
				log.Fatalln(err)
			}

			return
		case data := <-d.dataCh:
			// clear a terminal window
			if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				log.Fatalln(err)
			}
			// sync a terminal window
			if err := termbox.Sync(); err != nil {
				return
			}

			// set up a table header
			d.table.SetHeader(data.Header)

			// clear a table rows
			d.table.ClearRows()

			// set up a table rows
			for _, row := range data.Rows {
				d.table.Append(row)
			}

			// render a table
			d.table.Render()

			// draw a table
			if err := termbox.Flush(); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
