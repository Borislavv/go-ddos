package statservice

import (
	"github.com/Borislavv/go-ddos/internal/stat/domain/model"
	"sync"
	"time"
)

type Collector interface {
	Run(wg *sync.WaitGroup)

	Metric(stage int64) (metric *statmodel.Metrics, found bool)
	Stages() int64
	SummaryDuration() time.Duration
	SummaryRPS() int64

	AddWorker()
	Workers() int64
	RemoveWorker()

	AddTotalRequest()
	TotalRequests() int64
	SummaryTotalRequests() int64

	AddSuccessRequest()
	SuccessRequests() int64
	SummarySuccessRequests() int64

	AddFailedRequest()
	FailedRequests() int64
	SummaryFailedRequests() int64

	AddTotalRequestsDuration(d time.Duration)
	TotalRequestsDuration() (ms int64)
	AvgTotalRequestsDuration() time.Duration
	SummaryAvgTotalRequestsDuration() time.Duration

	AddSuccessRequestsDuration(d time.Duration)
	SuccessRequestsDuration() (ms int64)
	AvgSuccessRequestsDuration() time.Duration
	SummaryAvgSuccessRequestsDuration() time.Duration

	AddFailedRequestsDuration(d time.Duration)
	FailedRequestsDuration() (ms int64)
	AvgFailedRequestsDuration() time.Duration
	SummaryAvgFailedRequestsDuration() time.Duration

	HttpClientPoolBusy() int64
	SetHttpClientPoolBusy()
	HttpClientPoolTotal() int64
	SetHttpClientPoolTotal()
	HttpClientOutOfPool() int64
	SetHttpClientOutOfPool()
}
