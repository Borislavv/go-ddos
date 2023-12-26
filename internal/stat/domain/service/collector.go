package statservice

import (
	"ddos/config"
	"ddos/internal/stat/domain/model"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	mu                 *sync.RWMutex
	cfg                *config.Config
	durPerPercentile   time.Duration
	percentilesMetrics map[int64]*model.Metrics

	// time
	startedAt time.Time
	// workers
	workers int64
	// requests
	rps     int64
	total   int64
	success int64
	failed  int64
	// duration (ms)
	totalDuration   int64
	successDuration int64
	failedDuration  int64
}

func NewCollector(cfg *config.Config) *Collector {
	dur, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		log.Fatalln(err)
	}

	c := &Collector{
		mu:               &sync.RWMutex{},
		cfg:              cfg,
		durPerPercentile: time.Duration(math.Ceil(float64(dur.Nanoseconds() / cfg.Percentiles))),
	}

	if cfg.Percentiles <= 0 {
		atomic.CompareAndSwapInt64(&cfg.Percentiles, cfg.Percentiles, 1)
	}

	c.mu.Lock()
	c.percentilesMetrics = make(map[int64]*model.Metrics, cfg.Percentiles)
	c.mu.Unlock()

	return c
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
			metric.AddWorkers(prevMetric.Workers())
			metric.AddProcessors(prevMetric.Processors())
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

// Percentiles number. This value is not mutable.
func (c *Collector) Percentiles() int64 {
	return c.cfg.Percentiles
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
