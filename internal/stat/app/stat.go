package stat

import (
	"context"
	"ddos/config"
	displaymodel "ddos/internal/display/domain/model"
	displayservice "ddos/internal/display/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Stat struct {
	mu        *sync.RWMutex
	ctx       context.Context
	cfg       *config.Config
	renderer  *displayservice.Renderer
	collector *statservice.Collector
}

func New(
	ctx context.Context,
	cfg *config.Config,
	renderer *displayservice.Renderer,
	collector *statservice.Collector,
) *Stat {
	return &Stat{
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		cfg:       cfg,
		renderer:  renderer,
		collector: collector,
	}
}

func (s *Stat) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.sendStat(wg)
	wg.Wait()
}

func (s *Stat) sendStat(wg *sync.WaitGroup) {
	defer wg.Done()

	statTicker := time.NewTicker(time.Millisecond * 75)

	header := []string{
		"duration",
		"rps",
		"workers/processors",
		"total reqs.",
		"success reqs.",
		"failed reqs.",
		"avg. total reqs. dur.",
		"avg. success reqs. dur.",
		"avg. failed reqs. dur.",
		"stages",
		"goroutines",
	}

	rendererRows := make(map[int64][]string, s.cfg.Stages)

	for {
		select {
		case <-s.ctx.Done():
			var rows [][]string
			for percentile := int64(1); percentile <= s.collector.Stages(); percentile++ {
				row, ok := rendererRows[percentile]
				if ok {
					rows = append(rows, row)
					continue
				}

				metric, ok := s.collector.Metric(percentile)
				if !ok {
					continue
				}

				row = []string{
					metric.Duration().String(),
					fmt.Sprintf("%d", metric.RPS()),
					fmt.Sprintf("%d / %d", metric.Workers(), metric.Processors()),
					fmt.Sprintf("%d", metric.Total()),
					fmt.Sprintf("%d", metric.Success()),
					fmt.Sprintf("%d", metric.Failed()),
					metric.AvgTotalDuration().String(),
					metric.AvgSuccessDuration().String(),
					metric.AvgFailedDuration().String(),
					fmt.Sprintf("%d of %d", percentile, s.collector.Stages()),
					fmt.Sprintf("%d", runtime.NumGoroutine()),
				}

				rows = append(rows, row)
			}

			footer := []string{
				s.collector.SummaryDuration().String(),
				fmt.Sprintf("%d", s.collector.SummaryRPS()),
				fmt.Sprintf("%d / %d", s.collector.Workers(), s.collector.Processors()),
				fmt.Sprintf("%d", s.collector.SummaryTotal()),
				fmt.Sprintf("%d", s.collector.SummarySuccess()),
				fmt.Sprintf("%d", s.collector.SummaryFailed()),
				s.collector.SummaryAvgTotalDuration().String(),
				s.collector.SummaryAvgSuccessDuration().String(),
				s.collector.SummaryAvgFailedDuration().String(),
				"All",
				fmt.Sprintf("%d", runtime.NumGoroutine()),
			}

			s.renderer.SendSummary(
				&displaymodel.Table{
					Header: header,
					Rows:   rows,
					Footer: footer,
				},
			)
			return
		case <-statTicker.C:
			var rows [][]string
			for percentile := int64(1); percentile <= s.collector.Stages(); percentile++ {
				row, ok := rendererRows[percentile]
				if ok {
					rows = append(rows, row)
					continue
				}

				metric, ok := s.collector.Metric(percentile)
				if !ok {
					continue
				}

				row = []string{
					metric.Duration().String(),
					fmt.Sprintf("%d", metric.RPS()),
					fmt.Sprintf("%d / %d", metric.Workers(), metric.Processors()),
					fmt.Sprintf("%d", metric.Total()),
					fmt.Sprintf("%d", metric.Success()),
					fmt.Sprintf("%d", metric.Failed()),
					metric.AvgTotalDuration().String(),
					metric.AvgSuccessDuration().String(),
					metric.AvgFailedDuration().String(),
					fmt.Sprintf("%d of %d", percentile, s.collector.Stages()),
					fmt.Sprintf("%d", runtime.NumGoroutine()),
				}

				if metric.IsLocked() {
					rendererRows[percentile] = row
				}

				rows = append(rows, row)
			}

			s.renderer.SendData(
				&displaymodel.Table{
					Header: header,
					Rows:   rows,
				},
			)
		}
	}
}
