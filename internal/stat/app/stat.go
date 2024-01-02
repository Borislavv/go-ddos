package stat

import (
	"context"
	"github.com/Borislavv/go-ddos/config"
	displaymodel "github.com/Borislavv/go-ddos/internal/display/domain/model"
	displayservice "github.com/Borislavv/go-ddos/internal/display/domain/service"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"github.com/Borislavv/go-ddos/internal/stat/domain/model"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type App struct {
	ctx          context.Context
	logger       logservice.Logger
	renderer     displayservice.Renderer
	collector    statservice.Collector
	renderedRows map[int64][]string
}

func New(
	ctx context.Context,
	cfg *config.Config,
	logger logservice.Logger,
	renderer displayservice.Renderer,
	collector statservice.Collector,
) *App {
	return &App{
		ctx:          ctx,
		logger:       logger,
		renderer:     renderer,
		collector:    collector,
		renderedRows: make(map[int64][]string, cfg.Stages),
	}
}

func (s *App) Run(mwg *sync.WaitGroup) {
	tableCh := s.renderer.TableCh()
	summaryTableCh := s.renderer.SummaryTableCh()

	defer func() {
		s.logger.Println("stat.App.Run() is closed")
		_ = s.renderer.Close()
		mwg.Done()
	}()

	fps := time.NewTicker(time.Millisecond * 100)
	defer fps.Stop()

	for {
		select {
		case <-s.ctx.Done():
			table := &displaymodel.Table{
				Header: s.buildHeader(),
				Rows:   s.buildRows(),
				Footer: s.buildSummaryRow(),
			}

			select {
			case summaryTableCh <- table:
			default:
			}

			return
		case <-fps.C:
			table := &displaymodel.Table{
				Header: s.buildHeader(),
				Rows:   s.buildRows(),
			}

			select {
			case tableCh <- table:
			default:
			}
		}
	}
}

func (s *App) buildHeader() []string {
	return []string{
		"duration",
		"rps",
		"workers",
		"total reqs.",
		"success reqs.",
		"failed reqs.",
		"avg. total reqs. dur.",
		"avg. success reqs. dur.",
		"avg. failed reqs. dur.",
		"stages",
		"http pool",
		"goroutines",
	}
}

func (s *App) buildRows() [][]string {
	var rows [][]string
	for percentile := int64(1); percentile <= s.collector.Stages(); percentile++ {
		row, ok := s.renderedRows[percentile]
		if ok {
			rows = append(rows, row)
			continue
		}

		metric, ok := s.collector.Metric(percentile)
		if !ok {
			continue
		}

		row = s.buildRow(percentile, metric)

		if metric.IsLocked() {
			s.renderedRows[percentile] = row
		}

		rows = append(rows, row)
	}
	return rows
}

func (s *App) buildRow(percentile int64, metric *statmodel.Metrics) []string {
	return []string{
		metric.Duration().String(),
		strconv.FormatInt(metric.RPS(), 10),
		strconv.FormatInt(metric.Workers(), 10),
		strconv.FormatInt(metric.Total(), 10),
		strconv.FormatInt(metric.Success(), 10),
		strconv.FormatInt(metric.Failed(), 10),
		metric.AvgTotalDuration().String(),
		metric.AvgSuccessDuration().String(),
		metric.AvgFailedDuration().String(),
		strconv.FormatInt(percentile, 10) + " of " + strconv.FormatInt(s.collector.Stages(), 10),
		strconv.FormatInt(s.collector.HttpClientPoolBusy(), 10) + " / " + strconv.FormatInt(s.collector.HttpClientPoolTotal(), 10),
		strconv.Itoa(runtime.NumGoroutine()),
	}
}

func (s *App) buildSummaryRow() []string {
	return []string{
		s.collector.SummaryDuration().String(),
		strconv.FormatInt(s.collector.SummaryRPS(), 10),
		strconv.FormatInt(s.collector.Workers(), 10),
		strconv.FormatInt(s.collector.SummaryTotalRequests(), 10),
		strconv.FormatInt(s.collector.SummarySuccessRequests(), 10),
		strconv.FormatInt(s.collector.SummaryFailedRequests(), 10),
		s.collector.SummaryAvgTotalRequestsDuration().String(),
		s.collector.SummaryAvgSuccessRequestsDuration().String(),
		s.collector.SummaryAvgFailedRequestsDuration().String(),
		"All",
		strconv.FormatInt(s.collector.HttpClientPoolBusy(), 10) + " / " + strconv.FormatInt(s.collector.HttpClientPoolTotal(), 10),
		strconv.Itoa(runtime.NumGoroutine()),
	}
}
