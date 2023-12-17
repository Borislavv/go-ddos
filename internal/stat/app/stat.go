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
		"maxrps",
		"workers",
		"total reqs.",
		"success reqs.",
		"failed reqs.",
		"avg. total req. duration",
		"avg. success req. duration",
		"avg. failed req. duration",
	}

	row := []string{
		s.collector.Duration().String(),
		fmt.Sprintf("%d", s.collector.RPS()),
		fmt.Sprintf("%d", s.collector.Workers()),
		fmt.Sprintf("%d", s.collector.Total()),
		fmt.Sprintf("%d", s.collector.Success()),
		fmt.Sprintf("%d", s.collector.Failed()),
		s.collector.AvgTotalDuration().String(),
		s.collector.AvgSuccessDuration().String(),
		s.collector.AvgFailedDuration().String(),
	}

	for {
		select {
		case <-s.ctx.Done():
			s.summaryCh <- &displaymodel.Table{
				Header: header,
				Rows:   [][]string{row},
			}
			return
		case <-statTicker.C:
			s.dataCh <- &displaymodel.Table{
				Header: header,
				Rows:   [][]string{row},
			}
		}
	}
}
