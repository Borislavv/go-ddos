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
	rps = 0
)

type RendererV2Service struct {
	ctx       context.Context
	exitCh    chan<- os.Signal
	collector statservice.Collector

	logs *widgets.List
	rps  *widgets.Plot
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

	r.rps = r.initRpsPlot(width, height)
	r.logs = r.initLogsList(width, height)

	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	for {
		select {
		case e := <-ui.PollEvents():
			switch e.ID {
			case "<C-c>", "<C-z>":
				r.exitCh <- os.Interrupt
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				width = payload.Width

				r.rps.SetRect(0, 0, payload.Width, payload.Height-10)
				if len(r.rps.Data[rps]) >= width-10 {
					tmp := make([]float64, 0, len(r.rps.Data[rps]))
					for i := ((width - 10) / 100) * 10; i < width-10; i++ {
						tmp = append(tmp, r.rps.Data[rps][i])
					}
					r.rps.Data[rps] = tmp
				}

				r.logs.SetRect(0, payload.Height-10, payload.Width, payload.Height)

				ui.Clear()
			}
		case <-ticker.C:
			if len(r.rps.Data[rps]) >= width-10 {
				tmp := make([]float64, 0, len(r.rps.Data[rps]))
				for i := ((width - 10) / 100) * 10; i < width-10; i++ {
					tmp = append(tmp, r.rps.Data[rps][i])
				}
				r.rps.Data[rps] = tmp
			}
			r.rps.Data[0] = append(r.rps.Data[0], float64(r.collector.RPS()))

			// Проверяем, есть ли достаточно данных перед рендерингом
			if len(r.rps.Data[0]) > 1 {
				ui.Render(r.rps, r.logs)
			}
		}
	}
}

func (r *RendererV2Service) Write(p []byte) (n int, err error) {
	if r.logs == nil {
		return 0, nil
	}
	if len(r.logs.Rows) >= 10 {
		r.logs.Rows = r.logs.Rows[1:]
	}
	r.logs.Rows = append(r.logs.Rows, string(p))
	return len(p), nil
}

func (r *RendererV2Service) initRpsPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "RPS"

	plot.Data = make([][]float64, 1)         // need attention
	plot.Data[0] = make([]float64, 0, width) // need attention

	plot.AxesColor = ui.ColorWhite
	plot.LineColors[0] = ui.ColorCyan
	plot.SetRect(0, 0, width, height-5)
	return plot
}

func (r *RendererV2Service) initLogsList(width, height int) *widgets.List {
	logs := widgets.NewList()
	logs.Title = "Logs"

	logs.Rows = make([]string, 0, 10) // need attention

	logs.TextStyle = ui.NewStyle(ui.ColorYellow)
	logs.SetRect(0, height-10, width, height)
	return logs
}
