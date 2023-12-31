package statservice

import (
	"context"
	"ddos/internal/ddos/infrastructure/httpclient"
	"ddos/internal/stat/domain/model"
	"math"
	"sync"
	"time"
)

type Collector struct {
	ctx                context.Context
	mu                 *sync.RWMutex
	startedAt          time.Time
	httpClient         httpclient.Pooled
	durPerPercentile   time.Duration
	percentilesMetrics map[int64]*model.Metrics
	stages             int64
}

func NewCollector(ctx context.Context, httpClient httpclient.Pooled, duration time.Duration, stages int64) *Collector {
	c := &Collector{
		ctx:              ctx,
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

	c.percentilesMetrics = make(map[int64]*model.Metrics, c.stages)

	return c
}

func (c *Collector) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	refreshTicker := time.NewTicker(time.Millisecond * 100)
	defer refreshTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-refreshTicker.C:
			c.SetRPS()
			c.SetHttpClientPoolBusy()
			c.SetHttpClientPoolTotal()
		}
	}
}

func (c *Collector) firstMetric() *model.Metrics {
	metric, ok := c.Metric(1)
	if !ok {
		metric = model.NewMetric()
		c.mu.Lock()
		defer c.mu.Unlock()
		c.percentilesMetrics[1] = metric
	}
	return metric
}

func (c *Collector) currentMetric() *model.Metrics {
	current := int64(
		math.Round(float64(time.Since(c.startedAt).Milliseconds()/c.durPerPercentile.Milliseconds())),
	) + 1

	metric, ok := c.Metric(current)
	if !ok {
		metric = model.NewMetric()
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
func (c *Collector) Stages() int64 {
	return c.stages
}

func (c *Collector) Metric(stage int64) (metric *model.Metrics, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	metric, found = c.percentilesMetrics[stage]
	return metric, found
}

func (c *Collector) SummaryDuration() time.Duration {
	return time.Since(c.startedAt)
}

func (c *Collector) SetRPS() {
	c.currentMetric().SetRPS()
}

func (c *Collector) RPS() int64 {
	return c.currentMetric().RPS()
}

func (c *Collector) SummaryRPS() int64 {
	return int64(float64(c.summaryTotal()) / time.Since(c.startedAt).Seconds())
}

func (c *Collector) AddWorker() {
	c.currentMetric().AddWorkers(1)
}

func (c *Collector) RemoveWorker() {
	c.currentMetric().AddWorkers(-1)
}

func (c *Collector) Workers() int64 {
	return c.currentMetric().Workers()
}

func (c *Collector) AddTotal() {
	c.currentMetric().AddTotal()
}

func (c *Collector) Total() int64 {
	return c.currentMetric().Total()
}

func (c *Collector) summaryTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Total()
	}
	return t
}

func (c *Collector) SummaryTotal() int64 {
	return c.summaryTotal()
}

func (c *Collector) AddSuccess() {
	c.currentMetric().AddSuccess()
}

func (c *Collector) Success() int64 {
	return c.currentMetric().Success()
}

func (c *Collector) SummarySuccess() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Success()
	}
	return t
}

func (c *Collector) AddFailed() {
	c.currentMetric().AddFailed()
}

func (c *Collector) Failed() int64 {
	return c.currentMetric().Failed()
}

func (c *Collector) SummaryFailed() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Failed()
	}
	return t
}

// AddTotalDuration for current percentile.
func (c *Collector) AddTotalDuration(d time.Duration) {
	c.currentMetric().AddTotalDuration(d)
}

func (c *Collector) TotalDuration() time.Duration {
	return time.Duration(c.currentMetric().TotalDuration())
}

// AvgTotalDuration of current percentile.
func (c *Collector) AvgTotalDuration() time.Duration {
	return c.currentMetric().AvgTotalDuration()
}

// SummaryAvgTotalDuration of all percentiles.
func (c *Collector) SummaryAvgTotalDuration() time.Duration {
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

// AddSuccessDuration for current percentile.
func (c *Collector) AddSuccessDuration(d time.Duration) {
	c.currentMetric().AddSuccessDuration(d)
}

// SuccessDuration of current percentile.
func (c *Collector) SuccessDuration() (ms int64) {
	return c.currentMetric().SuccessDuration()
}

// AvgSuccessDuration of current percentile.
func (c *Collector) AvgSuccessDuration() time.Duration {
	return c.currentMetric().AvgSuccessDuration()
}

// SummaryAvgSuccessDuration of all percentiles.
func (c *Collector) SummaryAvgSuccessDuration() time.Duration {
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

// AddFailedDuration for current percentile.
func (c *Collector) AddFailedDuration(d time.Duration) {
	c.currentMetric().AddFailedDuration(d)
}

// FailedDuration of current percentile.
func (c *Collector) FailedDuration() (ms int64) {
	return c.currentMetric().FailedDuration()
}

// AvgFailedDuration of current percentile.
func (c *Collector) AvgFailedDuration() time.Duration {
	return c.currentMetric().AvgFailedDuration()
}

// SummaryAvgFailedDuration of all percentiles.
func (c *Collector) SummaryAvgFailedDuration() time.Duration {
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

func (c *Collector) HttpClientPoolBusy() int64 {
	return c.currentMetric().HttpClientPoolBusy()
}

func (c *Collector) SetHttpClientPoolBusy() {
	c.currentMetric().SetHttpClientPoolBusy(c.httpClient.Busy())
}

func (c *Collector) HttpClientPoolTotal() int64 {
	return c.currentMetric().HttpClientPoolTotal()
}

func (c *Collector) SetHttpClientPoolTotal() {
	c.currentMetric().SetHttpClientPoolTotal(c.httpClient.Total())
}
