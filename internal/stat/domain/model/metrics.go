package model

import "time"

type Metrics struct {
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
