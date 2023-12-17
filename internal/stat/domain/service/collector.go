package statservice

import (
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	mu *sync.RWMutex

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

func NewCollector() *Collector {
	return &Collector{
		mu: &sync.RWMutex{},
	}
}

func (c *Collector) SetStartedAt(s time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startedAt = s
}

func (c *Collector) StartedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.startedAt
}

func (c *Collector) Duration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.startedAt)
}

func (c *Collector) SetRPS() {
	atomic.CompareAndSwapInt64(
		&c.rps, atomic.LoadInt64(&c.rps), int64(float64(c.Total())/time.Since(c.StartedAt()).Seconds()),
	)
}

func (c *Collector) RPS() int64 {
	return atomic.LoadInt64(&c.rps)
}

func (c *Collector) AddWorker() {
	atomic.AddInt64(&c.workers, 1)
}

func (c *Collector) Workers() int64 {
	return atomic.LoadInt64(&c.workers)
}

func (c *Collector) AddTotal() {
	atomic.AddInt64(&c.total, 1)
}

func (c *Collector) Total() int64 {
	return atomic.LoadInt64(&c.total)
}

func (c *Collector) AddSuccess() {
	atomic.AddInt64(&c.success, 1)
}

func (c *Collector) Success() int64 {
	return atomic.LoadInt64(&c.success)
}

func (c *Collector) AddFailed() {
	atomic.AddInt64(&c.failed, 1)
}

func (c *Collector) Failed() int64 {
	return atomic.LoadInt64(&c.failed)
}

func (c *Collector) AddTotalDuration(d time.Duration) {
	atomic.AddInt64(&c.totalDuration, d.Milliseconds())
}

func (c *Collector) TotalDuration() (ms int64) {
	return atomic.LoadInt64(&c.totalDuration)
}

func (c *Collector) AvgTotalDuration() time.Duration {
	total := atomic.LoadInt64(&c.total)
	if total == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(atomic.LoadInt64(&c.totalDuration)/total) * time.Millisecond
	}
}

func (c *Collector) AddSuccessDuration(d time.Duration) {
	atomic.AddInt64(&c.successDuration, d.Milliseconds())
}

func (c *Collector) SuccessDuration() (ms int64) {
	return atomic.LoadInt64(&c.successDuration)
}

func (c *Collector) AvgSuccessDuration() time.Duration {
	success := atomic.LoadInt64(&c.success)
	if success == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(atomic.LoadInt64(&c.successDuration)/success) * time.Millisecond
	}
}

func (c *Collector) AddFailedDuration(d time.Duration) {
	atomic.AddInt64(&c.failedDuration, d.Milliseconds())
}

func (c *Collector) FailedDuration() (ms int64) {
	return atomic.LoadInt64(&c.failedDuration)
}

func (c *Collector) AvgFailedDuration() time.Duration {
	failed := atomic.LoadInt64(&c.failed)
	if failed == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(atomic.LoadInt64(&c.failedDuration)/failed) * time.Millisecond
	}
}
