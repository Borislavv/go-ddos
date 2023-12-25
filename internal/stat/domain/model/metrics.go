package model

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	mu *sync.RWMutex
	// state
	isMutable int64
	// time
	startedAt int64 // ms
	duration  int64 // ns
	// Workers
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

func NewMetric() *Metrics {
	return &Metrics{mu: &sync.RWMutex{}, isMutable: 1, startedAt: time.Now().UnixMilli()}
}

func (m *Metrics) Lock() {
	atomic.CompareAndSwapInt64(&m.isMutable, atomic.LoadInt64(&m.isMutable), 0)
}

func (m *Metrics) IsLocked() bool {
	return atomic.LoadInt64(&m.isMutable) == 0
}

func (m *Metrics) StartedAt() time.Time {
	return time.UnixMilli(atomic.LoadInt64(&m.startedAt))
}

func (m *Metrics) SetDuration() {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.CompareAndSwapInt64(
			&m.duration,
			atomic.LoadInt64(&m.duration),
			time.Since(time.UnixMilli(atomic.LoadInt64(&m.startedAt))).Nanoseconds(),
		)
	}
}

func (m *Metrics) Duration() time.Duration {
	m.SetDuration()
	return time.Duration(atomic.LoadInt64(&m.duration))
}

func (m *Metrics) Workers() int64 {
	return atomic.LoadInt64(&m.workers)
}

func (m *Metrics) AddWorkers(n int64) {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.workers, n)
	}
}

func (m *Metrics) SetRPS() {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.CompareAndSwapInt64(
			&m.rps,
			atomic.LoadInt64(&m.rps),
			int64(float64(atomic.LoadInt64(&m.total))/time.Since(time.UnixMilli(atomic.LoadInt64(&m.startedAt))).Seconds()),
		)
	}
}

func (m *Metrics) RPS() int64 {
	return atomic.LoadInt64(&m.rps)
}

func (m *Metrics) AddTotal() {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.total, 1)
	}
}

func (m *Metrics) Total() int64 {
	return atomic.LoadInt64(&m.total)
}

func (m *Metrics) AddSuccess() {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.success, 1)
	}
}

func (m *Metrics) Success() int64 {
	return atomic.LoadInt64(&m.success)
}

func (m *Metrics) AddFailed() {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.failed, 1)
	}
}

func (m *Metrics) Failed() int64 {
	return atomic.LoadInt64(&m.failed)
}

func (m *Metrics) AddTotalDuration(d time.Duration) {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.totalDuration, d.Milliseconds())
	}
}

func (m *Metrics) TotalDuration() int64 {
	return atomic.LoadInt64(&m.totalDuration)
}

func (m *Metrics) AvgTotalDuration() time.Duration {
	if m.Total() == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(m.TotalDuration()/m.Total()) * time.Millisecond
	}
}

func (m *Metrics) AddSuccessDuration(d time.Duration) {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.successDuration, d.Milliseconds())
	}
}

func (m *Metrics) SuccessDuration() int64 {
	return atomic.LoadInt64(&m.successDuration)
}

func (m *Metrics) AvgSuccessDuration() time.Duration {
	if m.Success() == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(m.SuccessDuration()/m.Success()) * time.Millisecond
	}
}

func (m *Metrics) AddFailedDuration(d time.Duration) {
	if atomic.LoadInt64(&m.isMutable) == 1 {
		atomic.AddInt64(&m.failedDuration, d.Milliseconds())
	}
}

func (m *Metrics) FailedDuration() int64 {
	return atomic.LoadInt64(&m.failedDuration)
}

func (m *Metrics) AvgFailedDuration() time.Duration {
	if m.Failed() == 0 {
		return time.Duration(0)
	} else {
		return time.Duration(m.FailedDuration()/m.Failed()) * time.Millisecond
	}
}
