package displayservice

import (
	"context"
	"fmt"
	"github.com/Borislavv/go-ddos/config"
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
	"strings"
	"sync"
	"time"
)

const (
	timestampDelimiter = " | "
	renderTickDur      = time.Millisecond * 100
)

const (
	widthThreshold  = 35
	heightThreshold = 15
)

var (
	totalDelimitersNumForMainCharts  = 0
	totalDelimitersNumForMinorCharts = 0
)

type RendererService struct {
	ctx       context.Context
	cfg       *config.Config
	exitCh    chan<- os.Signal
	collector statservice.Collector

	fd  *os.File
	log *widgets.List

	tst *widgets.Gauge
	inf *widgets.Paragraph

	rps       *widgets.Plot
	rpsTs     *widgets.Paragraph
	rpsTsData []string

	dur       *widgets.Plot
	durTs     *widgets.Paragraph
	durTsData []string

	grt       *widgets.Plot
	grtTs     *widgets.Paragraph
	grtTsData []string

	htp       *widgets.Plot
	htpTs     *widgets.Paragraph
	htpTsData []string

	wks       *widgets.Plot
	wksTs     *widgets.Paragraph
	wksTsData []string
}

func NewRendererService(
	ctx context.Context,
	cfg *config.Config,
	exitCh chan<- os.Signal,
	collector statservice.Collector,
) *RendererService {
	r := &RendererService{
		ctx:       ctx,
		cfg:       cfg,
		exitCh:    exitCh,
		collector: collector,
	}

	log.SetOutput(r)

	return r
}

func (r *RendererService) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	defer func() {
		if err := r.summary(); err != nil {
			panic(err)
		}
	}()

	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	width, height := ui.TerminalDimensions()
	if width < widthThreshold || height < heightThreshold {
		width, height = widthThreshold, heightThreshold
	}
	totalDelimitersNumForMainCharts = int(math.Round((((float64(width) / 100) * 59) - (2 * float64(len("15:04:05")))) / float64(len(timestampDelimiter))))
	totalDelimitersNumForMinorCharts = int(math.Round((((float64(width) / 100) * 40) - (2 * float64(len("15:04:05")))) / float64(len(timestampDelimiter))))

	r.log = r.initLogsList(width, height)
	r.tst = r.initDurationGauge(width, height)
	r.inf = r.initInfoParagraph(width, height)

	r.rps = r.initRpsPlot(width, height)
	r.rpsTs = r.initRpsTimestampXosParagraph(width, height)

	r.dur = r.initDurPlot(width, height)
	r.durTs = r.initDurTimestampXosParagraph(width, height)

	r.grt = r.initGoroutinesPlot(width, height)
	r.grtTs = r.initGoroutinesTimestampXosParagraph(width, height)

	r.htp = r.initHttpPoolPlot(width, height)
	r.htpTs = r.initHttpPoolTimestampXosParagraph(width, height)

	r.wks = r.initWorkersPlot(width, height)
	r.wksTs = r.initWorkersTimestampXosParagraph(width, height)

	ticker := time.NewTicker(renderTickDur)
	defer ticker.Stop()

	var rpsLineLey int
	var workersLineKey int
	var goroutinesLineLey int
	var httpPoolBusyClientsLineKey int
	var httpPoolOutOfPoolClientLineKey int
	var failedDurationLineKey int
	var successDurationLineKey int

	var isSatRPSLine bool
	var isSatWorkersLine bool
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
			case "<C-c>", "<C-z>", "q":
				r.exitCh <- os.Interrupt
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				if payload.Width < widthThreshold || payload.Height < heightThreshold {
					continue
				}

				width, height = payload.Width, payload.Height

				totalDelimitersNumForMainCharts = int(math.Round((((float64(width) / 100) * 59) - (2 * float64(len("15:04:05")))) / float64(len(timestampDelimiter))))
				totalDelimitersNumForMinorCharts = int(math.Round((((float64(width) / 100) * 40) - (2 * float64(len("15:04:05")))) / float64(len(timestampDelimiter))))

				// resize RPS chart
				r.rps.SetRect(
					0,
					0,
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
				)
				r.rpsTs.SetRect(
					0,
					int(math.Round((float64(height)/100)*34)),
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
				)
				r.renewRpsTimestampXosParagraph()

				// resize avg request duration char
				r.dur.SetRect(
					0,
					int(math.Round((float64(height)/100)*80)),
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
				)
				r.durTs.SetRect(
					0,
					int(math.Round((float64(height)/100)*74)),
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*80)),
				)
				r.renewDurTimestampXosParagraph()

				// resize goroutines chart
				r.grt.SetRect(
					int(math.Round((float64(width)/100)*60)),
					0,
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*20)),
				)
				r.grtTs.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*15)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*20)),
				)
				r.renewGoroutinesTimestampXosParagraph()

				// resize http client pool chart
				r.htp.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*20)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*40)),
				)
				r.htpTs.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*34)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*40)),
				)
				r.renewHttpPoolTimestampXosParagraph()

				// resize workers chart
				r.wks.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*40)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*60)),
				)
				r.wksTs.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*55)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*60)),
				)
				r.renewWorkersTimestampXosParagraph()

				// resize logs paragraph
				r.log.SetRect(
					0,
					int(math.Round((float64(height)/100)*80)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*100)),
				)

				// resize test duration gauge
				r.tst.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*60)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*70)),
				)

				// resize info paragraph
				r.inf.SetRect(
					int(math.Round((float64(width)/100)*60)),
					int(math.Round((float64(height)/100)*70)),
					int(math.Round((float64(width)/100)*100)),
					int(math.Round((float64(height)/100)*80)),
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

				r.rpsTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
				r.rpsTsData[len(r.rpsTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.rpsTsData) - 1)).Format("15:04:05")
				r.rpsTs.Text = strings.Join(r.rpsTsData, "")
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

				r.durTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
				r.durTsData[len(r.durTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.durTsData) - 1)).Format("15:04:05")
				r.durTs.Text = strings.Join(r.durTsData, "")
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

				r.durTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
				r.durTsData[len(r.durTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.durTsData) - 1)).Format("15:04:05")
				r.durTs.Text = strings.Join(r.durTsData, "")
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
			if isSatGoroutinesLine {
				if r.isMaxLenReachedForMinorPlots(width, r.grt.Data[goroutinesLineLey]) {
					r.grt.Data[goroutinesLineLey] = r.grt.Data[goroutinesLineLey][1:]
				}
				r.grt.Data[goroutinesLineLey] = append(r.grt.Data[goroutinesLineLey], g)

				if len(r.grtTsData) > 0 {
					r.grtTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
					r.grtTsData[len(r.grtTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.grtTsData) - 1)).Format("15:04:05")
				}
				r.grtTs.Text = strings.Join(r.grtTsData, "")
			}

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

				if len(r.htpTsData) > 0 {
					r.htpTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
					r.htpTsData[len(r.htpTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.htpTsData) - 1)).Format("15:04:05")
				}
				r.htpTs.Text = strings.Join(r.htpTsData, "")
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

				if len(r.htpTsData) > 0 {
					r.htpTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
					r.htpTsData[len(r.htpTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.htpTsData) - 1)).Format("15:04:05")
				}
				r.htpTs.Text = strings.Join(r.htpTsData, "")
			}

			// workers
			w := float64(r.collector.Workers())
			if w > 0 && !isSatWorkersLine {
				isSatWorkersLine = true

				r.wks.Data = append(r.wks.Data, make([]float64, 0, (width/100)*40))
				workersLineKey = len(r.wks.Data) - 1
				r.wks.Data[workersLineKey] = append(r.wks.Data[workersLineKey], 0)
				r.wks.LineColors[workersLineKey] = ui.ColorRed
			}
			if isSatWorkersLine {
				if r.isMaxLenReachedForMinorPlots(width, r.wks.Data[workersLineKey]) {
					r.wks.Data[workersLineKey] = r.wks.Data[workersLineKey][1:]
				}
				r.wks.Data[workersLineKey] = append(r.wks.Data[workersLineKey], w)

				if len(r.wksTsData) > 0 {
					r.wksTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
					r.wksTsData[len(r.wksTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.wksTsData) - 1)).Format("15:04:05")
				}
				r.wksTs.Text = strings.Join(r.wksTsData, "")
			}

			r.tst.Percent = int(time.Since(r.collector.StartedAt()).Milliseconds() / (r.cfg.DurationValue / 100).Milliseconds())
			r.tst.Label = fmt.Sprintf(
				"%v%% (remaining time %v)",
				time.Since(r.collector.StartedAt()).Milliseconds()/(r.cfg.DurationValue.Milliseconds()/100),
				strings.Split((r.cfg.DurationValue - time.Since(r.collector.StartedAt())).String(), ".")[0]+"s",
			)

			var items = []ui.Drawable{r.tst, r.inf, r.log}
			if len(r.rps.Data) > 0 {
				items = append(items, r.rps)
				items = append(items, r.rpsTs)
			}
			if len(r.dur.Data) > 0 {
				items = append(items, r.dur)
				items = append(items, r.durTs)
			}
			if len(r.grt.Data) > 0 {
				items = append(items, r.grt)
				items = append(items, r.grtTs)
			}
			if len(r.htp.Data) > 0 {
				items = append(items, r.htp)
				items = append(items, r.htpTs)
			}
			if len(r.wks.Data) > 0 {
				items = append(items, r.wks)
				items = append(items, r.wksTs)
			}

			ui.Render(items...)
		}
	}
}

func (r *RendererService) isMaxLenReachedForMinorPlots(width int, data []float64) bool {
	return len(data) >= int(math.Round((float64(width)/100)*38))
}

func (r *RendererService) isMaxLenReachedForMainPlots(width int, data []float64) bool {
	return len(data) >= int(math.Round((float64(width)/100)*57))
}

func (r *RendererService) Write(p []byte) (n int, err error) {
	if r.log == nil {
		return 0, nil
	}
	if len(r.log.Rows) >= 10 {
		r.log.Rows = r.log.Rows[1:]
	}
	r.log.Rows = append(r.log.Rows, string(p))
	_, _ = r.fd.Write(p)
	return len(p), nil
}

func (r *RendererService) initRpsPlot(width, height int) *widgets.Plot {
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

func (r *RendererService) initRpsTimestampXosParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Border = true
	p.WrapText = true
	p.TextStyle.Fg = ui.ColorYellow

	r.rpsTsData = make([]string, 0, totalDelimitersNumForMainCharts+2)
	for j := 0; j < cap(r.rpsTsData)-1; j++ {
		r.rpsTsData = append(r.rpsTsData, timestampDelimiter)
	}

	p.SetRect(
		0,
		int(math.Round((float64(height)/100)*34)),
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*40)),
	)
	return p
}

func (r *RendererService) renewRpsTimestampXosParagraph() {
	r.rpsTsData = make([]string, 0, totalDelimitersNumForMainCharts+2)
	for j := 0; j < cap(r.rpsTsData)-1; j++ {
		r.rpsTsData = append(r.rpsTsData, timestampDelimiter)
	}
	r.rpsTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
	r.rpsTsData[len(r.rpsTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.rpsTsData) - 1)).Format("15:04:05")
	r.rpsTs.Text = strings.Join(r.rpsTsData, "")
}

func (r *RendererService) initDurPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "Duration"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 0, 2)
	plot.LineColors = make([]ui.Color, 2)
	plot.LineColors[0] = ui.ColorGreen
	plot.LineColors[1] = ui.ColorRed

	plot.SetRect(
		0,
		int(math.Round((float64(height)/100)*80)),
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*40)),
	)

	return plot
}

func (r *RendererService) initDurTimestampXosParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Border = true
	p.WrapText = true
	p.TextStyle.Fg = ui.ColorYellow

	r.durTsData = make([]string, 0, totalDelimitersNumForMainCharts+2)
	for j := 0; j < cap(r.durTsData)-1; j++ {
		r.durTsData = append(r.durTsData, timestampDelimiter)
	}

	p.SetRect(
		0,
		int(math.Round((float64(height)/100)*74)),
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*80)),
	)
	return p
}

func (r *RendererService) renewDurTimestampXosParagraph() {
	r.durTsData = make([]string, 0, totalDelimitersNumForMainCharts+2)
	for j := 0; j < cap(r.durTsData)-1; j++ {
		r.durTsData = append(r.durTsData, timestampDelimiter)
	}
	r.durTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
	r.durTsData[len(r.durTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.durTsData) - 1)).Format("15:04:05")
	r.durTs.Text = strings.Join(r.durTsData, "")
}

func (r *RendererService) initGoroutinesPlot(width, height int) *widgets.Plot {
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

func (r *RendererService) initGoroutinesTimestampXosParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Border = true
	p.WrapText = true
	p.TextStyle.Fg = ui.ColorYellow

	r.grtTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.grtTsData)-1; j++ {
		r.grtTsData = append(r.grtTsData, timestampDelimiter)
	}

	p.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*15)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*20)),
	)
	return p
}

func (r *RendererService) renewGoroutinesTimestampXosParagraph() {
	r.grtTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.grtTsData)-1; j++ {
		r.grtTsData = append(r.grtTsData, timestampDelimiter)
	}
	if len(r.grtTsData) > 0 {
		r.grtTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
		r.grtTsData[len(r.grtTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.grtTsData) - 1)).Format("15:04:05")
	}
	r.grtTs.Text = strings.Join(r.grtTsData, "")
}

func (r *RendererService) initHttpPoolPlot(width, height int) *widgets.Plot {
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

func (r *RendererService) initHttpPoolTimestampXosParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Border = true
	p.WrapText = true
	p.TextStyle.Fg = ui.ColorYellow

	r.htpTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.htpTsData)-1; j++ {
		r.htpTsData = append(r.htpTsData, timestampDelimiter)
	}

	p.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*34)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*40)),
	)
	return p
}

func (r *RendererService) renewHttpPoolTimestampXosParagraph() {
	r.htpTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.htpTsData)-1; j++ {
		r.htpTsData = append(r.htpTsData, timestampDelimiter)
	}
	if len(r.htpTsData) > 0 {
		r.htpTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
		r.htpTsData[len(r.htpTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.htpTsData) - 1)).Format("15:04:05")
	}
	r.htpTs.Text = strings.Join(r.htpTsData, "")
}

func (r *RendererService) initWorkersPlot(width, height int) *widgets.Plot {
	plot := widgets.NewPlot()
	plot.Title = "Workers"
	plot.AxesColor = ui.ColorWhite

	plot.Data = make([][]float64, 0, 1)

	plot.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*40)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*60)),
	)

	return plot
}

func (r *RendererService) initWorkersTimestampXosParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Border = true
	p.WrapText = true
	p.TextStyle.Fg = ui.ColorYellow

	r.wksTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.wksTsData)-1; j++ {
		r.wksTsData = append(r.wksTsData, timestampDelimiter)
	}

	p.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*55)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*60)),
	)
	return p
}

func (r *RendererService) renewWorkersTimestampXosParagraph() {
	r.wksTsData = make([]string, 0, totalDelimitersNumForMinorCharts+2)
	for j := 0; j < cap(r.wksTsData)-1; j++ {
		r.wksTsData = append(r.wksTsData, timestampDelimiter)
	}
	if len(r.wksTsData) > 0 {
		r.wksTsData[0] = time.Now().Add(renderTickDur).Format("15:04:05")
		r.wksTsData[len(r.wksTsData)-1] = time.Now().Add(time.Duration(int(renderTickDur)*len(r.wksTsData) - 1)).Format("15:04:05")
	}
	r.wksTs.Text = strings.Join(r.wksTsData, "")
}

func (r *RendererService) initLogsList(width, height int) *widgets.List {
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

func (r *RendererService) initDurationGauge(width, height int) *widgets.Gauge {
	g := widgets.NewGauge()
	g.Title = "Test duration"
	g.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*60)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*70)),
	)

	g.Percent = int(time.Since(r.collector.StartedAt()).Milliseconds() / (r.cfg.DurationValue / 100).Milliseconds())
	g.Label = fmt.Sprintf(
		"%v%% (remeining time %v)",
		time.Since(r.collector.StartedAt())/(r.cfg.DurationValue/100),
		r.cfg.DurationValue-time.Since(r.collector.StartedAt()),
	)

	return g
}

func (r *RendererService) initInfoParagraph(width, height int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Title = "Info"
	p.Text = "Press [CTRL+C/Z](fg:red,mode:bold) or [q](fg:red,mode:bold) for exit.\n\n" +
		"[Green line](fg:green,mode:bold) is a positive values, " +
		"[red line](fg:red,mode:bold) is a negative."
	p.SetRect(
		int(math.Round((float64(width)/100)*60)),
		int(math.Round((float64(height)/100)*70)),
		int(math.Round((float64(width)/100)*100)),
		int(math.Round((float64(height)/100)*80)),
	)
	p.BorderStyle.Fg = ui.ColorYellow
	return p
}

func (r *RendererService) summary() error {
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
