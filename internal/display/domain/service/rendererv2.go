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
	widthThreshold  = 35
	heightThreshold = 15
)

const (
	rps = 0

	durSuccess = 0
	durFailed  = 1

	goroutines = 0

	httpPoolBusy      = 0
	httpPoolOutOfPool = 1
)

type RendererV2Service struct {
	ctx       context.Context
	exitCh    chan<- os.Signal
	collector statservice.Collector

	log *widgets.List
	dur *widgets.Plot
	rps *widgets.Plot
	grt *widgets.Plot
	htp *widgets.Plot
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
	if width < widthThreshold || height < heightThreshold {
		width, height = widthThreshold, heightThreshold
	}

	r.dur = r.initDurPlot(width, height)
	r.log = r.initLogsList(width, height)
	r.grt = r.initGoroutinesPlot(width, height)
	r.rps = r.initRpsPlot(width, height)
	r.htp = r.initHttpPoolPlot(width, height)

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	var rpsLineLey int
	var goroutinesLineLey int
	var httpPoolBusyClientsLineKey int
	var httpPoolOutOfPoolClientLineKey int
	var failedDurationLineKey int
	var successDurationLineKey int

	var isSatRPSLine bool
	var isSatGoroutinesLine bool
	var isSatHttpPoolBusyClientsLine bool
	var isSatHttpOutOfPoolClientsLine bool
	var isSatFailedDurationLine bool
	var isSatSuccessDurationLine bool

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
				if payload.Width < widthThreshold || payload.Height < heightThreshold {
					_, _ = fmt.Fprintf(
						r, "warning: minimum size [%dx%d] was reached", widthThreshold, heightThreshold,
					)
					continue
				}

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
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*20)),
				)

				r.htp.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*20)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*40)),
				)

				r.log.SetRect(
					0,
					int(math.Round((float64(height)/100)*80)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*100)),
				)

				ui.Clear()
			}
		case <-ticker.C:
			// avg rps number
			p := float64(r.collector.RPS())
			if p > 0 && !isSatRPSLine {
				isSatRPSLine = true

				r.rps.Data = append(r.rps.Data, make([]float64, 0, int(math.Round((float64(width)/100)*60))))
				rpsLineLey = len(r.rps.Data) - 1
				r.rps.Data[rpsLineLey] = append(r.rps.Data[rpsLineLey], 0)
				r.rps.LineColors[rpsLineLey] = ui.ColorGreen

			}
			if isSatRPSLine {
				if r.isMaxLenReachedForMainPlots(width, r.rps.Data[rpsLineLey]) {
					r.rps.Data[rpsLineLey] = r.rps.Data[rpsLineLey][1:]
				}
				r.rps.Data[rpsLineLey] = append(r.rps.Data[rpsLineLey], p)
			}

			// avg success requests duration
			s := float64(r.collector.AvgSuccessRequestsDuration().Milliseconds())
			if s > 0 && !isSatSuccessDurationLine {
				isSatSuccessDurationLine = true

				r.dur.Data = append(r.dur.Data, make([]float64, 0, (width/100)*60))
				successDurationLineKey = len(r.dur.Data) - 1
				r.dur.Data[successDurationLineKey] = append(r.dur.Data[successDurationLineKey], 0)
				r.dur.LineColors[successDurationLineKey] = ui.ColorGreen
			}
			if isSatSuccessDurationLine {
				if r.isMaxLenReachedForMainPlots(width, r.dur.Data[successDurationLineKey]) {
					r.dur.Data[successDurationLineKey] = r.dur.Data[successDurationLineKey][1:]
				}
				r.dur.Data[successDurationLineKey] = append(r.dur.Data[successDurationLineKey], s)
			}

			// avg failed requests duration
			f := float64(r.collector.AvgFailedRequestsDuration().Milliseconds())
			if f > 0 && !isSatFailedDurationLine {
				isSatFailedDurationLine = true

				r.dur.Data = append(r.dur.Data, make([]float64, 0, (width/100)*60))
				failedDurationLineKey = len(r.dur.Data) - 1
				r.dur.Data[failedDurationLineKey] = append(r.dur.Data[failedDurationLineKey], 0)
				r.dur.LineColors[failedDurationLineKey] = ui.ColorRed
			}
			if isSatFailedDurationLine {
				if r.isMaxLenReachedForMainPlots(width, r.dur.Data[failedDurationLineKey]) {
					r.dur.Data[failedDurationLineKey] = r.dur.Data[failedDurationLineKey][1:]
				}
				r.dur.Data[failedDurationLineKey] = append(r.dur.Data[failedDurationLineKey], f)
			}

			// number of goroutines
			g := float64(runtime.NumGoroutine())
			if g > 0 && !isSatGoroutinesLine {
				isSatGoroutinesLine = true

				r.grt.Data = append(r.grt.Data, make([]float64, 0, (width/100)*40))
				goroutinesLineLey = len(r.grt.Data) - 1
				r.grt.Data[goroutinesLineLey] = append(r.grt.Data[goroutinesLineLey], 0)
				r.grt.LineColors[goroutinesLineLey] = ui.ColorGreen
			}
			if r.isMaxLenReachedForMinorPlots(width, r.grt.Data[goroutinesLineLey]) {
				r.grt.Data[goroutinesLineLey] = r.grt.Data[goroutinesLineLey][1:]
			}
			r.grt.Data[goroutinesLineLey] = append(r.grt.Data[goroutinesLineLey], g)

			// busy http clients into the pool
			b := float64(r.collector.HttpClientPoolBusy())
			if b > 0 && !isSatHttpPoolBusyClientsLine {
				isSatHttpPoolBusyClientsLine = true

				r.htp.Data = append(r.htp.Data, make([]float64, 0, (width/100)*40))
				httpPoolBusyClientsLineKey = len(r.htp.Data) - 1
				r.htp.Data[httpPoolBusyClientsLineKey] = append(r.htp.Data[httpPoolBusyClientsLineKey], 0)
				r.htp.LineColors[httpPoolBusyClientsLineKey] = ui.ColorGreen
			}
			if isSatHttpPoolBusyClientsLine {
				if r.isMaxLenReachedForMinorPlots(width, r.htp.Data[httpPoolBusyClientsLineKey]) {
					r.htp.Data[httpPoolBusyClientsLineKey] = r.htp.Data[httpPoolBusyClientsLineKey][1:]
				}
				r.htp.Data[httpPoolBusyClientsLineKey] = append(r.htp.Data[httpPoolBusyClientsLineKey], b)
			}

			// http client which out of pool (extra clients)
			o := float64(r.collector.HttpClientOutOfPool())
			if o > 0 && !isSatHttpOutOfPoolClientsLine {
				isSatHttpOutOfPoolClientsLine = true

				r.htp.Data = append(r.htp.Data, make([]float64, 0, (width/100)*40))
				httpPoolOutOfPoolClientLineKey = len(r.htp.Data) - 1
				r.htp.Data[httpPoolOutOfPoolClientLineKey] = append(r.htp.Data[httpPoolOutOfPoolClientLineKey], 0)
				r.htp.LineColors[httpPoolOutOfPoolClientLineKey] = ui.ColorRed
			}
			if isSatHttpOutOfPoolClientsLine {
				if r.isMaxLenReachedForMinorPlots(width, r.htp.Data[httpPoolOutOfPoolClientLineKey]) {
					r.htp.Data[httpPoolOutOfPoolClientLineKey] = r.htp.Data[httpPoolOutOfPoolClientLineKey][1:]
				}
				r.htp.Data[httpPoolOutOfPoolClientLineKey] = append(r.htp.Data[httpPoolOutOfPoolClientLineKey], o)
			}

			var items = []ui.Drawable{r.log}
			if len(r.rps.Data) > 0 {
				items = append(items, r.rps)
			}
			if len(r.dur.Data) > 0 {
				items = append(items, r.dur)
			}
			if len(r.grt.Data) > 0 {
				items = append(items, r.grt)
			}
			if len(r.htp.Data) > 0 {
				items = append(items, r.htp)
			}

			ui.Render(items...)
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

	plot.Data = make([][]float64, 0, 1)

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

	plot.Data = make([][]float64, 0, 2)

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

	plot.Data = make([][]float64, 0, 1)

	plot.SetRect(
		int(math.Round((float64(width)/100)*60)),
		0,
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*20)),
	)

	return plot
}

func (r *RendererV2Service) initHttpPoolPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "HttpPool"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 0, 2)

	plot.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*20)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*40)),
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
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*100)),
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
