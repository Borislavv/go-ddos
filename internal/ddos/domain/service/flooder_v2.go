package ddosservice

import (
	"context"
	ddos "ddos/config"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type FlooderV2 struct {
	mu  *sync.RWMutex
	ctx context.Context

	cfg       *ddos.Config
	logger    *logservice.Logger
	collector *statservice.Collector

	spawnRequestsWorkerCh chan struct{}
	closeRequestsWorkerCh chan struct{}
	exitRequestsWorkersCh chan struct{}

	spawnResponsesProcessorCh chan struct{}
	closeResponsesProcessorCh chan struct{}
	exitResponsesProcessorsCh chan struct{}
}

func NewFlooderV2(
	ctx context.Context,
	cfg *ddos.Config,
	logger *logservice.Logger,
	collector *statservice.Collector,
) *FlooderV2 {
	return &FlooderV2{
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		cfg:       cfg,
		logger:    logger,
		collector: collector,
	}
}

func (f *FlooderV2) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	wg2 := &sync.WaitGroup{}
	defer wg2.Wait()
	wg2.Add(1)
	go f.requestsWorkersBalancer(wg2)

	wg3 := &sync.WaitGroup{}
	defer wg3.Wait()
	wg3.Add(1)
	go f.responsesProcessorsBalancer(wg3)
}

func (f *FlooderV2) requestsWorkersBalancer(wg *sync.WaitGroup) {
	defer wg.Done()

	spawnTicker := time.NewTicker(time.Millisecond * 100)
	defer spawnTicker.Stop()

	requestSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.cfg.MaxRPS)*1.10))
	defer requestSendTicker.Stop()

	select {
	case <-f.ctx.Done():
		// broadcasting exit event for all request workers
		close(f.exitRequestsWorkersCh)
		return
	case <-spawnTicker.C:
		wg.Add(1)
		go f.spawnRequestsWorker(wg, requestSendTicker)
	}
}

func (f *FlooderV2) responsesProcessorsBalancer(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	select {
	case <-f.ctx.Done():
		// broadcasting exit event for all response processors
		close(f.exitResponsesProcessorsCh)
		return
	case <-ticker.C:
		wg.Add(1)
	}
}

func (f *FlooderV2) spawnRequestsWorker(wg *sync.WaitGroup, sendRequestTicker *time.Ticker) {
	defer wg.Done()

	for {
		select {
		case <-f.exitRequestsWorkersCh:
			return
		case <-f.closeRequestsWorkerCh:
			return
		case <-sendRequestTicker.C:
			f.sendRequest()
		}
	}
}

func (f *FlooderV2) sendRequest() {
	rand.Seed(time.Now().UnixNano())

	s := time.Now()
	defer f.collector.AddTotalDuration(time.Since(s))

	resp, err := http.Get(fmt.Sprintf("%v&ts=%d", f.cfg.URL, rand.Uint64()))
	if err != nil || resp.StatusCode != 200 {
		f.collector.AddFailedDuration(time.Since(s))
	} else {
		f.collector.AddSuccessDuration(time.Since(s))
	}

	//f.sendResponseOnProcessing(resp, err)
}
