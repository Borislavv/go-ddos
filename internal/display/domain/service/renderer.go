package displayservice

import (
	"context"
	displaymodel "ddos/internal/display/domain/model"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
)

type Renderer struct {
	ctx       context.Context
	dataCh    <-chan *displaymodel.Table
	summaryCh <-chan *displaymodel.Table
	exitCh    chan<- os.Signal
	stopCh    chan struct{}
}

func NewRenderer(
	ctx context.Context,
	dataCh <-chan *displaymodel.Table,
	summaryCh <-chan *displaymodel.Table,
	exitCh chan<- os.Signal,
) *Renderer {
	return &Renderer{
		ctx:       ctx,
		exitCh:    exitCh,
		dataCh:    dataCh,
		summaryCh: summaryCh,
		stopCh:    make(chan struct{}, 1),
	}
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
	defer wg.Done()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	table := tablewriter.NewWriter(os.Stdout)

	for {
		select {
		case <-r.stopCh:
			termbox.Close()

			r.exitCh <- os.Interrupt

			data := <-r.summaryCh

			// draw the summary table
			table = tablewriter.NewWriter(os.Stdout)

			// draw a header of the summary tables
			table.SetHeader(data.Header)

			// draw rows of the summary table
			for _, row := range data.Rows {
				table.Append(row)
			}

			// render the summary table
			table.Render()

			// draw the summary table
			if err = termbox.Flush(); err != nil {
				log.Fatalln(err)
			}

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

			// set up a table header
			table.SetHeader(data.Header)

			// clear a table rows
			table.ClearRows()

			// set up a table rows
			for _, row := range data.Rows {
				table.Append(row)
			}

			// render a table
			table.Render()

			// draw a table
			if err = termbox.Flush(); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
