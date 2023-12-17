package displayservice

import (
	"context"
	model "ddos/internal/display/domain/model"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
)

type Renderer struct {
	ctx       context.Context
	dataCh    chan *model.Table
	summaryCh chan *model.Table
	exitCh    chan os.Signal
	stopCh    chan struct{}
}

func NewRenderer(ctx context.Context, exitCh chan os.Signal, dataCh chan *model.Table, summaryCh chan *model.Table) *Renderer {
	return &Renderer{
		ctx:       ctx,
		exitCh:    exitCh,
		dataCh:    dataCh,
		summaryCh: summaryCh,
		stopCh:    make(chan struct{}, 1),
	}
}

func (d *Renderer) Close(wg *sync.WaitGroup) {
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

func (d *Renderer) Draw(wg *sync.WaitGroup) {
	defer wg.Done()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	table := tablewriter.NewWriter(os.Stdout)

	for {
		select {
		case <-d.stopCh:
			termbox.Close()

			d.exitCh <- os.Interrupt

			log.Println("renderer: send interrupt signal")

			summary := <-d.summaryCh

			log.Println("renderer: received summary data")

			// draw the summary table
			table = tablewriter.NewWriter(os.Stdout)

			// draw a header of the summary tables
			table.SetHeader(summary.Header)

			// draw rows of the summary table
			for _, row := range summary.Rows {
				table.Append(row)
			}

			// render the summary table
			table.Render()

			// draw the summary table
			if err = termbox.Flush(); err != nil {
				log.Fatalln(err)
			}

			return
		case data := <-d.dataCh:
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
