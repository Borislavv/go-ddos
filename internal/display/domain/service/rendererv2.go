package displayservice

import (
	"context"
	"fmt"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	rps        = 0
	durSuccess = 0
	durFailed  = 1
	goroutines = 0
)

type RendererV2Service struct {
	ctx       context.Context
	exitCh    chan<- os.Signal
	collector statservice.Collector

	log *widgets.List
	dur *widgets.Plot
	rps *widgets.Plot
	grt *widgets.Plot
}

func NewRendererV2Service(
	ctx context.Context,
	exitCh chan<- os.Signal,
	collector statservice.Collector,
) *RendererV2Service {
	return &RendererV2Service{
		ctx:       ctx,
		exitCh:    exitCh,
		collector: collector,
	}
}

func (r *RendererV2Service) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	defer func() {
		if err := r.summary(); err != nil {
			log.Println(err)
		}
	}()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	width, height := ui.TerminalDimensions()

	r.dur = r.initDurPlot(width, height)
	r.log = r.initLogsList(width, height)
	r.grt = r.initGoroutinesPlot(width, height)
	r.rps = r.initRpsPlot(width, height)

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	eventCh := ui.PollEvents()
	for {
		select {
		case <-r.ctx.Done():
			return
		case e := <-eventCh:
			switch e.ID {
			case "<C-c>", "<C-z>":
				r.exitCh <- os.Interrupt
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				width, height = payload.Width, payload.Height

				r.rps.SetRect(
					0,
					0,
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
				)

				r.dur.SetRect(
					0,
					int(math.Round((float64(height)/100)*80)),
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
				)

				r.grt.SetRect(
					int(math.Round((float64(width)/100)*60)),
					0,
					width,
					int(math.Round((float64(height)/100)*20)),
				)

				r.log.SetRect(
					0,
					int(math.Round((float64(height)/100)*80)),
					width,
					height,
				)

				ui.Clear()
			}
		case <-ticker.C:
			// avg rps number
			if r.isMaxLenReachedForMainPlots(width, r.rps.Data[rps]) {
				r.rps.Data[rps] = r.rps.Data[rps][1:]
			}
			r.rps.Data[rps] = append(r.rps.Data[rps], float64(r.collector.RPS()))

			// avg success requests duration
			if r.isMaxLenReachedForMainPlots(width, r.dur.Data[durSuccess]) {
				r.dur.Data[durSuccess] = r.dur.Data[durSuccess][1:]
			}
			r.dur.Data[durSuccess] = append(r.dur.Data[durSuccess], float64(r.collector.AvgSuccessRequestsDuration().Milliseconds()))

			// avg failed requests duration
			if r.isMaxLenReachedForMainPlots(width, r.dur.Data[durFailed]) {
				r.dur.Data[durFailed] = r.dur.Data[durFailed][1:]
			}
			f := float64(r.collector.AvgFailedRequestsDuration().Milliseconds())
			if f == 0 {
				r.dur.LineColors[durFailed] = ui.ColorBlack
			} else {
				r.dur.LineColors[durFailed] = ui.ColorRed
			}
			r.dur.Data[durFailed] = append(r.dur.Data[durFailed], f)

			// number of goroutines
			if r.isMaxLenReachedForMinorPlots(width, r.grt.Data[goroutines]) {
				r.grt.Data[goroutines] = r.grt.Data[goroutines][1:]
			}
			r.grt.Data[goroutines] = append(r.grt.Data[goroutines], float64(runtime.NumGoroutine()))

			ui.Render(r.rps, r.dur, r.grt, r.log)
		}
	}
}

func (r *RendererV2Service) isMaxLenReachedForMinorPlots(width int, data []float64) bool {
	return len(data) >= int(math.Round((float64(width)/100)*38))
}

func (r *RendererV2Service) isMaxLenReachedForMainPlots(width int, data []float64) bool {
	return len(data) >= int(math.Round((float64(width)/100)*57))
}

func (r *RendererV2Service) Write(p []byte) (n int, err error) {
	if r.log == nil {
		return 0, nil
	}
	if len(r.log.Rows) >= 10 {
		r.log.Rows = r.log.Rows[1:]
	}
	r.log.Rows = append(r.log.Rows, string(p))
	return len(p), nil
}

func (r *RendererV2Service) initRpsPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()

	plot.Title = "RPS"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 1)
	plot.Data[rps] = make([]float64, 0, int(math.Round((float64(width)/100)*60)))
	plot.Data[rps] = append(plot.Data[rps], 0)
	plot.LineColors[rps] = ui.ColorGreen

	plot.SetRect(
		0,
		0,
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*40)),
	)

	return plot
}

func (r *RendererV2Service) initDurPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "Duration"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 2)

	plot.Data[durSuccess] = make([]float64, 0, (width/100)*60)
	plot.Data[durFailed] = make([]float64, 0, (width/100)*60)

	plot.Data[durSuccess] = append(plot.Data[durSuccess], 0)
	plot.Data[durFailed] = append(plot.Data[durFailed], 0)

	plot.LineColors[durSuccess] = ui.ColorGreen
	plot.LineColors[durFailed] = ui.ColorRed

	plot.SetRect(
		0,
		int(math.Round((float64(height)/100)*80)),
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*40)),
	)

	return plot
}

func (r *RendererV2Service) initGoroutinesPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "Goroutines"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 1)
	plot.Data[goroutines] = make([]float64, 0, (width/100)*40)
	plot.Data[goroutines] = append(plot.Data[goroutines], 0)
	plot.LineColors[goroutines] = ui.ColorGreen

	plot.SetRect(
		int(math.Round((float64(width)/100)*60)),
		0,
		width,
		int(math.Round((float64(height)/100)*20)),
	)

	return plot
}

func (r *RendererV2Service) initLogsList(width, height int) *widgets.List {
	logs := widgets.NewList()
	logs.Title = "Logs"

	logs.Rows = make([]string, 0, 10)

	logs.TextStyle = ui.NewStyle(ui.ColorYellow)
	logs.SetRect(
		0,
		int(math.Round((float64(height)/100)*80)),
		width,
		height,
	)
	return logs
}

func (r *RendererV2Service) summary() error {
	// reinitialized the table as the summary
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	// set up a table header
	table.SetHeader([]string{
		"duration",
		"rps",
		"workers",
		"total reqs.",
		"success reqs.",
		"failed reqs.",
		"avg. total reqs. dur.",
		"avg. success reqs. dur.",
		"avg. failed reqs. dur.",
		"http pool",
		"goroutines",
	})

	// clear a table rows
	table.ClearRows()

	var outside string
	if r.collector.HttpClientOutOfPool() > 0 {
		outside = fmt.Sprintf(" (OUTSIDE: %d)", r.collector.HttpClientOutOfPool())
	}

	// set up a table rows
	table.Append([]string{
		r.collector.SummaryDuration().String(),
		strconv.FormatInt(r.collector.SummaryRPS(), 10),
		strconv.FormatInt(r.collector.Workers(), 10),
		strconv.FormatInt(r.collector.SummaryTotalRequests(), 10),
		strconv.FormatInt(r.collector.SummarySuccessRequests(), 10),
		strconv.FormatInt(r.collector.SummaryFailedRequests(), 10),
		r.collector.SummaryAvgTotalRequestsDuration().String(),
		r.collector.SummaryAvgSuccessRequestsDuration().String(),
		r.collector.SummaryAvgFailedRequestsDuration().String(),
		strconv.FormatInt(r.collector.HttpClientPoolBusy(), 10) + " / " +
			strconv.FormatInt(r.collector.HttpClientPoolTotal(), 10) + outside,
		strconv.Itoa(runtime.NumGoroutine()),
	})

	// render a table
	table.Render()

	// draw a table
	if err := termbox.Flush(); err != nil {
		return err
	}

	return nil
}
