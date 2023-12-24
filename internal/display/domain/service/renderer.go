package displayservice

import (
	"context"
	"ddos/config"
	displaymodel "ddos/internal/display/domain/model"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
)

type Renderer struct {
	ctx       context.Context
	cfg       *config.Config
	dataCh    chan *displaymodel.Table
	summaryCh chan *displaymodel.Table
	exitCh    chan<- os.Signal
	stopCh    chan struct{}
}

func NewRenderer(ctx context.Context, cfg *config.Config, exitCh chan<- os.Signal) *Renderer {
	return &Renderer{
		ctx:       ctx,
		exitCh:    exitCh,
		dataCh:    make(chan *displaymodel.Table, cfg.MaxRPS),
		summaryCh: make(chan *displaymodel.Table),
		stopCh:    make(chan struct{}, 1),
	}
}

func (r *Renderer) SendData(data *displaymodel.Table) {
	r.dataCh <- data
}

func (r *Renderer) SendSummary(data *displaymodel.Table) {
	r.summaryCh <- data
}

func (r *Renderer) Close(wg *sync.WaitGroup) {
	defer func() {
		close(r.stopCh)
		wg.Done()
	}()

	for {
		select {
		case <-r.ctx.Done():
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

func (r *Renderer) Draw(wg *sync.WaitGroup) {
	defer func() {
		close(r.dataCh)
		close(r.summaryCh)
		wg.Done()
	}()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	for {
		select {
		case <-r.stopCh:
			termbox.Close()

			r.exitCh <- os.Interrupt

			data := <-r.summaryCh

			// reinitialized the table as the summary
			table = tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_CENTER)

			// draw the summary table
			r.draw(table, data)

			return
		case data := <-r.dataCh:
			// clear a terminal window
			if err = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				log.Fatalln(err)
			}
			// sync a terminal window
			if err = termbox.Sync(); err != nil {
				return
			}

			// draw a table
			r.draw(table, data)
		}
	}
}

func (r *Renderer) draw(table *tablewriter.Table, data *displaymodel.Table) {
	// set up a table header
	table.SetHeader(data.Header)

	// clear a table rows
	table.ClearRows()

	// set up a table rows
	table.AppendBulk(data.Rows)

	if data.Footer != nil {
		table.SetFooter(data.Footer)
	}

	// render a table
	table.Render()

	// draw a table
	if err := termbox.Flush(); err != nil {
		log.Fatalln(err)
	}
}
