package displayservice

import (
	"context"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"sync"
	"time"
)

const (
	rps        = 0
	durSuccess = 0
	durFailed  = 1
)

type RendererV2Service struct {
	ctx       context.Context
	exitCh    chan<- os.Signal
	collector statservice.Collector

	log *widgets.List
	dur *widgets.Plot
	rps *widgets.Plot
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

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	width, height := ui.TerminalDimensions()

	r.dur = r.initDurPlot(width, height)
	r.rps = r.initRpsPlot(width, height)
	r.log = r.initLogsList(width, height)

	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case e := <-ui.PollEvents():
			switch e.ID {
			case "<C-c>", "<C-z>":
				r.exitCh <- os.Interrupt
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				r.rps.SetRect(0, 0, payload.Width, payload.Height-30)
				r.dur.SetRect(0, payload.Height-30, payload.Width, payload.Height-10)
				r.log.SetRect(0, payload.Height-10, payload.Width, payload.Height)
				ui.Clear()
			}
		case <-ticker.C:
			if r.isMaxLenReached(width, r.rps.Data[rps]) {
				r.rps.Data[rps] = r.reduceDataSlice(r.rps.Data[rps])
			}
			r.rps.Data[rps] = append(r.rps.Data[rps], float64(r.collector.RPS()))

			if r.isMaxLenReached(width, r.dur.Data[durSuccess]) {
				r.dur.Data[durSuccess] = r.reduceDataSlice(r.dur.Data[durSuccess])
			}
			r.dur.Data[durSuccess] = append(r.dur.Data[durSuccess], float64(r.collector.AvgSuccessRequestsDuration().Milliseconds()))

			if r.isMaxLenReached(width, r.dur.Data[durFailed]) {
				r.dur.Data[durFailed] = r.reduceDataSlice(r.dur.Data[durFailed])
			}
			r.dur.Data[durFailed] = append(r.dur.Data[durFailed], float64(r.collector.AvgFailedRequestsDuration().Milliseconds()))
			r.dur.DataLabels = append(r.dur.DataLabels, "asdasd")

			var items []ui.Drawable
			if len(r.rps.Data[rps]) > 1 {
				items = append(items, r.rps)
			}
			if len(r.dur.Data[durSuccess]) > 1 && len(r.dur.Data[durFailed]) > 1 {
				items = append(items, r.dur)
			}
			items = append(items, r.log)

			ui.Render(items...)
		}
	}
}

func (r *RendererV2Service) isMaxLenReached(width int, data []float64) bool {
	return len(data) >= ((width / 100) * 98)
}

func (r *RendererV2Service) reduceDataSlice(old []float64) (new []float64) {
	return old[1:]
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
	plot.Data[rps] = make([]float64, 0, width)
	plot.LineColors[rps] = ui.ColorGreen

	plot.SetRect(0, 0, width, height-30)

	return plot
}

func (r *RendererV2Service) initDurPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "Duration"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 2)

	plot.Data[durSuccess] = make([]float64, 0, width)
	plot.Data[durFailed] = make([]float64, 0, width)

	plot.LineColors[durSuccess] = ui.ColorGreen
	plot.LineColors[durFailed] = ui.ColorRed

	plot.SetRect(0, height-30, width, height-10)

	return plot
}

func (r *RendererV2Service) initLogsList(width, height int) *widgets.List {
	logs := widgets.NewList()
	logs.Title = "Logs"

	logs.Rows = make([]string, 0, 10)

	logs.TextStyle = ui.NewStyle(ui.ColorYellow)
	logs.SetRect(0, height-10, width, height)
	return logs
}
