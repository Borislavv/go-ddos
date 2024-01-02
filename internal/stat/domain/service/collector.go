package statservice

import (
	"context"
	"ddos/internal/flooder/infrastructure/httpclient"
	logservice "ddos/internal/log/domain/service"
	"ddos/internal/stat/domain/model"
	"math"
	"sync"
	"time"
)

type CollectorService struct {
	ctx context.Context
	mu  *sync.RWMutex

	logger     logservice.Logger
	httpClient httpclient.Pooled

	startedAt          time.Time
	durPerPercentile   time.Duration
	percentilesMetrics map[int64]*statmodel.Metrics
	stages             int64
}

func NewCollectorService(
	ctx context.Context,
	logger logservice.Logger,
	httpClient httpclient.Pooled,
	duration time.Duration,
	stages int64,
) *CollectorService {
	c := &CollectorService{
		ctx:              ctx,
		logger:           logger,
		startedAt:        time.Now(),
		httpClient:       httpClient,
		mu:               &sync.RWMutex{},
		durPerPercentile: time.Duration(math.Ceil(float64(duration.Nanoseconds() / stages))),
	}

	if stages <= 0 {
		c.stages = 1
	} else {
		c.stages = stages
	}

	c.percentilesMetrics = make(map[int64]*statmodel.Metrics, c.stages)

	return c
}

func (c *CollectorService) Run(wg *sync.WaitGroup) {
	defer func() {
		c.logger.Println("stat.CollectorService.Run() is closed")
		wg.Done()
	}()

	refreshTicker := time.NewTicker(time.Millisecond * 100)
	defer refreshTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-refreshTicker.C:
			c.setRPS()
			c.SetHttpClientPoolBusy()
			c.SetHttpClientPoolTotal()
		}
	}
}

func (c *CollectorService) firstMetric() *statmodel.Metrics {
	metric, ok := c.Metric(1)
	if !ok {
		metric = statmodel.NewMetric()
		c.mu.Lock()
		defer c.mu.Unlock()
		c.percentilesMetrics[1] = metric
	}
	return metric
}

func (c *CollectorService) currentMetric() *statmodel.Metrics {
	current := int64(
		math.Round(float64(time.Since(c.startedAt).Milliseconds()/c.durPerPercentile.Milliseconds())),
	) + 1

	metric, ok := c.Metric(current)
	if !ok {
		metric = statmodel.NewMetric()
		if prevMetric, isset := c.Metric(current - 1); isset {
			prevMetric.Lock()
			metric = prevMetric.Clone()
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		c.percentilesMetrics[current] = metric
	}
	return metric
}

// Stages number. This value is not mutable.
func (c *CollectorService) Stages() int64 {
	return c.stages
}

func (c *CollectorService) Metric(stage int64) (metric *statmodel.Metrics, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	metric, found = c.percentilesMetrics[stage]
	return metric, found
}

func (c *CollectorService) SummaryDuration() time.Duration {
	return time.Since(c.startedAt)
}

func (c *CollectorService) setRPS() {
	c.currentMetric().SetRPS()
}

func (c *CollectorService) SummaryRPS() int64 {
	return int64(float64(c.summaryTotal()) / time.Since(c.startedAt).Seconds())
}

func (c *CollectorService) AddWorker() {
	c.currentMetric().AddWorkers(1)
}

func (c *CollectorService) RemoveWorker() {
	c.currentMetric().AddWorkers(-1)
}

func (c *CollectorService) Workers() int64 {
	return c.currentMetric().Workers()
}

func (c *CollectorService) AddTotalRequest() {
	c.currentMetric().AddTotal()
}

func (c *CollectorService) TotalRequests() int64 {
	return c.currentMetric().Total()
}

func (c *CollectorService) summaryTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Total()
	}
	return t
}

func (c *CollectorService) SummaryTotalRequests() int64 {
	return c.summaryTotal()
}

func (c *CollectorService) AddSuccessRequest() {
	c.currentMetric().AddSuccess()
}

func (c *CollectorService) SuccessRequests() int64 {
	return c.currentMetric().Success()
}

func (c *CollectorService) SummarySuccessRequests() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Success()
	}
	return t
}

func (c *CollectorService) AddFailedRequest() {
	c.currentMetric().AddFailed()
}

func (c *CollectorService) FailedRequests() int64 {
	return c.currentMetric().Failed()
}

func (c *CollectorService) SummaryFailedRequests() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Failed()
	}
	return t
}

// AddTotalRequestsDuration for current percentile.
func (c *CollectorService) AddTotalRequestsDuration(d time.Duration) {
	c.currentMetric().AddTotalDuration(d)
}

func (c *CollectorService) TotalRequestsDuration() (ms int64) {
	return c.currentMetric().TotalDuration()
}

// AvgTotalRequestsDuration of current percentile.
func (c *CollectorService) AvgTotalRequestsDuration() time.Duration {
	return c.currentMetric().AvgTotalDuration()
}

// SummaryAvgTotalRequestsDuration of all percentiles.
func (c *CollectorService) SummaryAvgTotalRequestsDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total int64
	var totalDuration int64
	for _, metric := range c.percentilesMetrics {
		total += metric.Total()
		totalDuration += metric.TotalDuration()
	}

	if total == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(totalDuration/total) * time.Millisecond
	}
}

// AddSuccessRequestsDuration for current percentile.
func (c *CollectorService) AddSuccessRequestsDuration(d time.Duration) {
	c.currentMetric().AddSuccessDuration(d)
}

// SuccessRequestsDuration of current percentile.
func (c *CollectorService) SuccessRequestsDuration() (ms int64) {
	return c.currentMetric().SuccessDuration()
}

// AvgSuccessRequestsDuration of current percentile.
func (c *CollectorService) AvgSuccessRequestsDuration() time.Duration {
	return c.currentMetric().AvgSuccessDuration()
}

// SummaryAvgSuccessRequestsDuration of all percentiles.
func (c *CollectorService) SummaryAvgSuccessRequestsDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var success int64
	var successDuration int64
	for _, metric := range c.percentilesMetrics {
		success += metric.Success()
		successDuration += metric.SuccessDuration()
	}

	if success == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(successDuration/success) * time.Millisecond
	}
}

// AddFailedRequestsDuration for current percentile.
func (c *CollectorService) AddFailedRequestsDuration(d time.Duration) {
	c.currentMetric().AddFailedDuration(d)
}

// FailedRequestsDuration of current percentile.
func (c *CollectorService) FailedRequestsDuration() (ms int64) {
	return c.currentMetric().FailedDuration()
}

// AvgFailedRequestsDuration of current percentile.
func (c *CollectorService) AvgFailedRequestsDuration() time.Duration {
	return c.currentMetric().AvgFailedDuration()
}

// SummaryAvgFailedRequestsDuration of all percentiles.
func (c *CollectorService) SummaryAvgFailedRequestsDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var failed int64
	var failedDuration int64
	for _, metric := range c.percentilesMetrics {
		failed += metric.Failed()
		failedDuration += metric.FailedDuration()
	}

	if failed == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(failedDuration/failed) * time.Millisecond
	}
}

func (c *CollectorService) HttpClientPoolBusy() int64 {
	return c.currentMetric().HttpClientPoolBusy()
}

func (c *CollectorService) SetHttpClientPoolBusy() {
	c.currentMetric().SetHttpClientPoolBusy(c.httpClient.Busy())
}

func (c *CollectorService) HttpClientPoolTotal() int64 {
	return c.currentMetric().HttpClientPoolTotal()
}

func (c *CollectorService) SetHttpClientPoolTotal() {
	c.currentMetric().SetHttpClientPoolTotal(c.httpClient.Total())
}
