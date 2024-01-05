package stat

import (
	"context"
	"ddos/config"
	displaymodel "ddos/internal/display/domain/model"
	displayservice "ddos/internal/display/domain/service"
	logservice "ddos/internal/log/domain/service"
	"ddos/internal/stat/domain/model"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"runtime"
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
		fmt.Sprintf("%d", metric.RPS()),
		fmt.Sprintf("%d", metric.Workers()),
		fmt.Sprintf("%d", metric.Total()),
		fmt.Sprintf("%d", metric.Success()),
		fmt.Sprintf("%d", metric.Failed()),
		metric.AvgTotalDuration().String(),
		metric.AvgSuccessDuration().String(),
		metric.AvgFailedDuration().String(),
		fmt.Sprintf("%d of %d", percentile, s.collector.Stages()),
		fmt.Sprintf("%d / %d", s.collector.HttpClientPoolBusy(), s.collector.HttpClientPoolTotal()),
		fmt.Sprintf("%d", runtime.NumGoroutine()),
	}
}

func (s *App) buildSummaryRow() []string {
	return []string{
		s.collector.SummaryDuration().String(),
		fmt.Sprintf("%d", s.collector.SummaryRPS()),
		fmt.Sprintf("%d", s.collector.Workers()),
		fmt.Sprintf("%d", s.collector.SummaryTotalRequests()),
		fmt.Sprintf("%d", s.collector.SummarySuccessRequests()),
		fmt.Sprintf("%d", s.collector.SummaryFailedRequests()),
		s.collector.SummaryAvgTotalRequestsDuration().String(),
		s.collector.SummaryAvgSuccessRequestsDuration().String(),
		s.collector.SummaryAvgFailedRequestsDuration().String(),
		"All",
		fmt.Sprintf("%d / %d", s.collector.HttpClientPoolBusy(), s.collector.HttpClientPoolTotal()),
		fmt.Sprintf("%d", runtime.NumGoroutine()),
	}
}
