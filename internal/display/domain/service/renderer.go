package displayservice

import (
	"context"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
	"time"
)

type Renderer struct {
	ctx       context.Context
	collector *statservice.Collector
	exitCh    chan os.Signal
	stopCh    chan struct{}
}

func NewRenderer(
	ctx context.Context,
	collector *statservice.Collector,
	exitCh chan os.Signal,
) *Renderer {
	return &Renderer{
		ctx:       ctx,
		exitCh:    exitCh,
		collector: collector,
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

	renderTicker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-r.stopCh:
			termbox.Close()

			r.exitCh <- os.Interrupt

			// draw the summary table
			table = tablewriter.NewWriter(os.Stdout)

			// draw a header of the summary tables
			table.SetHeader([]string{
				"duration",
				"maxrps",
				"workers",
				"total reqs.",
				"success reqs.",
				"failed reqs.",
				"avg. total req. duration",
				"avg. success req. duration",
				"avg. failed req. duration",
			})

			// draw rows of the summary table
			table.Append([]string{
				r.collector.Duration().String(),
				fmt.Sprintf("%d", r.collector.RPS()),
				fmt.Sprintf("%d", r.collector.Workers()),
				fmt.Sprintf("%d", r.collector.Total()),
				fmt.Sprintf("%d", r.collector.Success()),
				fmt.Sprintf("%d", r.collector.Failed()),
				r.collector.AvgTotalDuration().String(),
				r.collector.AvgSuccessDuration().String(),
				r.collector.AvgFailedDuration().String(),
			})

			// render the summary table
			table.Render()

			// draw the summary table
			if err = termbox.Flush(); err != nil {
				log.Fatalln(err)
			}

			return
		case <-renderTicker.C:
			// clear a terminal window
			if err = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				log.Fatalln(err)
			}
			// sync a terminal window
			if err = termbox.Sync(); err != nil {
				return
			}

			// set up a table header
			table.SetHeader([]string{
				"duration",
				"maxrps",
				"workers",
				"total reqs.",
				"success reqs.",
				"failed reqs.",
				"avg. total req. duration",
				"avg. success req. duration",
				"avg. failed req. duration",
			})

			// clear a table rows
			table.ClearRows()

			// set up a table rows
			table.Append([]string{
				r.collector.Duration().String(),
				fmt.Sprintf("%d", r.collector.RPS()),
				fmt.Sprintf("%d", r.collector.Workers()),
				fmt.Sprintf("%d", r.collector.Total()),
				fmt.Sprintf("%d", r.collector.Success()),
				fmt.Sprintf("%d", r.collector.Failed()),
				r.collector.AvgTotalDuration().String(),
				r.collector.AvgSuccessDuration().String(),
				r.collector.AvgFailedDuration().String(),
			})

			// render a table
			table.Render()

			// draw a table
			if err = termbox.Flush(); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
