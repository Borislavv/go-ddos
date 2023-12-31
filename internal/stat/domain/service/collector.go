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
	httpClient         httpclient.Pooled
	durPerPercentile   time.Duration
	percentilesMetrics map[int64]*model.Metrics
	stages             int64
}

func NewCollector(ctx context.Context, httpClient httpclient.Pooled, duration time.Duration, stages int64) *Collector {
	c := &Collector{
		ctx:              ctx,
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

func (c *Collector) HttpClientPoolBusy() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().HttpClientPoolBusy()
}

func (c *Collector) SetHttpClientPoolBusy() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().SetHttpClientPoolBusy(c.httpClient.Busy())
}

func (c *Collector) HttpClientPoolTotal() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().HttpClientPoolTotal()
}

func (c *Collector) SetHttpClientPoolTotal() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().SetHttpClientPoolTotal(c.httpClient.Total())
}

func (c *Collector) CurrentPercentile() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentPercentile()
}

func (c *Collector) currentPercentile() int64 {
	return int64(
		math.Round(float64(time.Since(c.firstMetric().StartedAt()).Milliseconds()/c.durPerPercentile.Milliseconds())),
	) + 1
}

func (c *Collector) firstPercentile() int64 {
	return 1
}

func (c *Collector) currentMetric() *model.Metrics {
	metric, ok := c.percentilesMetrics[c.currentPercentile()]
	if !ok {
		metric = model.NewMetric()

		if prevMetric, isset := c.percentilesMetrics[c.currentPercentile()-1]; isset {
			prevMetric.Lock()
			metric = prevMetric.Clone()
		}

		c.percentilesMetrics[c.currentPercentile()] = metric
	}
	return metric
}

func (c *Collector) firstMetric() *model.Metrics {
	metric, ok := c.percentilesMetrics[c.firstPercentile()]
	if !ok {
		metric = model.NewMetric()
		c.percentilesMetrics[c.firstPercentile()] = metric
	}
	return metric
}

// Stages number. This value is not mutable.
func (c *Collector) Stages() int64 {
	return c.stages
}

func (c *Collector) Metric(percentile int64) (metric *model.Metrics, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	metric, found = c.percentilesMetrics[percentile]
	return metric, found
}

func (c *Collector) SummaryDuration() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	return time.Since(c.firstMetric().StartedAt())
}

func (c *Collector) SetRPS() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentMetric().SetRPS()
}

func (c *Collector) RPS() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().RPS()
}

func (c *Collector) SummaryRPS() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return int64(float64(c.summaryTotal()) / time.Since(c.firstMetric().StartedAt()).Seconds())
}

func (c *Collector) AddWorker() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddWorkers(1)
}

func (c *Collector) RemoveWorker() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddWorkers(-1)
}

func (c *Collector) Workers() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().Workers()
}

func (c *Collector) AddProcessor() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddProcessors(1)
}

func (c *Collector) RemoveProcessor() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddProcessors(-1)
}

func (c *Collector) Processors() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().Processors()
}

func (c *Collector) AddTotal() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddTotal()
}

func (c *Collector) Total() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().Total()
}

func (c *Collector) summaryTotal() int64 {
	var t int64
	for _, metric := range c.percentilesMetrics {
		t += metric.Total()
	}
	return t
}

func (c *Collector) SummaryTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.summaryTotal()
}

func (c *Collector) AddSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddSuccess()
}

func (c *Collector) Success() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddFailed()
}

func (c *Collector) Failed() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddTotalDuration(d)
}

func (c *Collector) TotalDuration() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	return time.Duration(c.currentMetric().TotalDuration())
}

// AvgTotalDuration of current percentile.
func (c *Collector) AvgTotalDuration() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddSuccessDuration(d)

}

// SuccessDuration of current percentile.
func (c *Collector) SuccessDuration() (ms int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().SuccessDuration()
}

// AvgSuccessDuration of current percentile.
func (c *Collector) AvgSuccessDuration() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetric().AddFailedDuration(d)
}

// FailedDuration of current percentile.
func (c *Collector) FailedDuration() (ms int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentMetric().FailedDuration()
}

// AvgFailedDuration of current percentile.
func (c *Collector) AvgFailedDuration() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

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
