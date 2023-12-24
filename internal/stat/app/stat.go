package stat

import (
	"context"
	displaymodel "ddos/internal/display/domain/model"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"sync"
	"time"
)

type Stat struct {
	mu        *sync.RWMutex
	ctx       context.Context
	dataCh    chan<- *displaymodel.Table
	summaryCh chan<- *displaymodel.Table
	collector *statservice.Collector
}

func New(
	ctx context.Context,
	dataCh chan<- *displaymodel.Table,
	summaryCh chan<- *displaymodel.Table,
	collector *statservice.Collector,
) *Stat {
	return &Stat{
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		dataCh:    dataCh,
		summaryCh: summaryCh,
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

	statTicker := time.NewTicker(time.Millisecond * 100)

	header := []string{
		"duration",
		"rps",
		"workers",
		"total reqs.",
		"success reqs.",
		"failed reqs.",
		"avg. total req. duration",
		"avg. success req. duration",
		"avg. failed req. duration",
		"current percentile",
	}
	
	for {
		select {
		case <-s.ctx.Done():
			var rows [][]string
			for percentile := int64(1); percentile <= s.collector.Percentiles(); percentile++ {
				metric, ok := s.collector.Metric(percentile)
				if !ok {
					continue
				}

				rows = append(rows, []string{
					metric.Duration().String(),
					fmt.Sprintf("%d", metric.RPS()),
					fmt.Sprintf("%d", metric.Workers()),
					fmt.Sprintf("%d", metric.Total()),
					fmt.Sprintf("%d", metric.Success()),
					fmt.Sprintf("%d", metric.Failed()),
					metric.AvgTotalDuration().String(),
					metric.AvgSuccessDuration().String(),
					metric.AvgFailedDuration().String(),
					fmt.Sprintf("%d of %d", percentile, s.collector.Percentiles()),
				})
			}

			rows = append(rows, []string{
				s.collector.SummaryDuration().String(),
				fmt.Sprintf("%d", s.collector.SummaryRPS()),
				fmt.Sprintf("%d", s.collector.Workers()),
				fmt.Sprintf("%d", s.collector.SummaryTotal()),
				fmt.Sprintf("%d", s.collector.SummarySuccess()),
				fmt.Sprintf("%d", s.collector.SummaryFailed()),
				s.collector.SummaryAvgTotalDuration().String(),
				s.collector.SummaryAvgSuccessDuration().String(),
				s.collector.SummaryAvgFailedDuration().String(),
				"All",
			})

			s.summaryCh <- &displaymodel.Table{
				Header: header,
				Rows:   rows,
			}
			return
		case <-statTicker.C:
			var rows [][]string
			for percentile := int64(1); percentile <= s.collector.Percentiles(); percentile++ {
				metric, ok := s.collector.Metric(percentile)
				if !ok {
					continue
				}

				rows = append(rows, []string{
					metric.Duration().String(),
					fmt.Sprintf("%d", metric.RPS()),
					fmt.Sprintf("%d", metric.Workers()),
					fmt.Sprintf("%d", metric.Total()),
					fmt.Sprintf("%d", metric.Success()),
					fmt.Sprintf("%d", metric.Failed()),
					metric.AvgTotalDuration().String(),
					metric.AvgSuccessDuration().String(),
					metric.AvgFailedDuration().String(),
					fmt.Sprintf("%d of %d", percentile, s.collector.Percentiles()),
				})
			}

			s.dataCh <- &displaymodel.Table{
				Header: header,
				Rows:   rows,
			}
		}
	}
}
