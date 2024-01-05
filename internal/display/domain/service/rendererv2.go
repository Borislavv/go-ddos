package displayservice

import (
	"context"
	"fmt"
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

	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Sprintf("e: %v, width: %d, height: %d, len: %d, cap: %d", e, width, height, len(r.rps.Data[rps]), cap(r.rps.Data[rps])))
		}
	}()

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
				height = payload.Height

				r.rps.SetRect(0, 0, width, height-10)
				log.Printf("width: %d, len(rps): %d, ((width / 100) * 95): %d", width, len(r.rps.Data[rps]), (width/100)*95)
				if len(r.rps.Data[rps]) >= ((width / 100) * 98) {
					tmp := make([]float64, 0, width)
					for _, v := range r.rps.Data[rps] {
						tmp = append(tmp, v)
					}
					r.rps.Data[rps] = tmp
				}

				r.logs.SetRect(0, height-10, width, height)

				ui.Clear()
			}
		case <-ticker.C:
			log.Printf("width: %d, len(rps): %d, ((width / 100) * 95): %d", width, len(r.rps.Data[rps]), (width/100)*95)
			if len(r.rps.Data[rps]) >= ((width / 100) * 98) {
				log.Println("OUT OF RANGE")
				tmp := make([]float64, 0, width)
				for i := width - ((width / 100) * 95); i < len(r.rps.Data[rps]); i++ {
					tmp = append(tmp, r.rps.Data[rps][i])
				}
				r.rps.Data[rps] = tmp
			}
			r.rps.Data[rps] = append(r.rps.Data[rps], float64(r.collector.RPS()))

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

	plot.Data = make([][]float64, 1)           // need attention
	plot.Data[rps] = make([]float64, 0, width) // need attention
	plot.DataLabels = make([]string, 0, width)

	plot.AxesColor = ui.ColorWhite
	plot.LineColors[rps] = ui.ColorGreen
	plot.SetRect(0, 0, width, height-10)
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
