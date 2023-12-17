package ddosservice

import (
	"context"
	display "ddos/internal/display/app"
	displaymodel "ddos/internal/display/domain/model"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Flooder struct {
	mu  *sync.RWMutex
	ctx context.Context

	// dependencies
	display *display.Display

	// settings
	rps    int   // number of max rps
	maxwks int64 // number of max active workers

	// metrics
	actwks  int64     // number of active workers
	total   int64     // number of total reqs.
	success int64     // number of success reqs.
	failed  int64     // number of failed reqs.
	started time.Time // time of the ddos started

	totalDurationMs   int64 // accumulative duration of total reqs.
	successDurationMs int64 // accumulative duration of success reqs.
	failedDurationMs  int64 // accumulative duration of failed reqs.
}

func NewFlooder(ctx context.Context, rps int, maxwks int64, display *display.Display) *Flooder {
	return &Flooder{
		mu:      &sync.RWMutex{},
		ctx:     ctx,
		rps:     rps,
		maxwks:  maxwks,
		display: display,
	}
}

func (f *Flooder) Run() {
	f.started = time.Now()

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		requestSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.rps)*1.10))
		defer requestSendTicker.Stop()

		threadSpawnTicker := time.NewTicker(time.Millisecond * 50)
		defer threadSpawnTicker.Stop()

		for {
			crps := int(float64(atomic.LoadInt64(&f.total)) / time.Since(f.started).Seconds())

			select {
			case <-f.ctx.Done():
				f.sendStat(crps, true)
				return
			case <-threadSpawnTicker.C:
				f.spawnThread(wg, requestSendTicker, crps)
			default:
				f.sendStat(crps, false)
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
}

func (f *Flooder) sendStat(crps int, isSummary bool) {
	total := atomic.LoadInt64(&f.total)

	var avgTotalDur string
	if total == 0 {
		avgTotalDur = time.Duration(0).String()
	} else {
		avgTotalDur = ((time.Duration(atomic.LoadInt64(&f.totalDurationMs) / total)) * time.Millisecond).String()
	}

	success := atomic.LoadInt64(&f.success)

	var avgSuccessDur string
	if success == 0 {
		avgSuccessDur = time.Duration(0).String()
	} else {
		avgSuccessDur = ((time.Duration(atomic.LoadInt64(&f.successDurationMs) / success)) * time.Millisecond).String()
	}

	failed := atomic.LoadInt64(&f.failed)

	var avgFailedDur string
	if failed == 0 {
		avgFailedDur = time.Duration(0).String()
	} else {
		avgFailedDur = ((time.Duration(atomic.LoadInt64(&f.failedDurationMs) / failed)) * time.Millisecond).String()
	}

	t := &displaymodel.Table{
		Header: []string{
			"duration",
			"rps",
			"workers",
			"total reqs.",
			"success reqs.",
			"failed reqs.",
			"avg. total req. duration",
			"avg. success req. duration",
			"avg. failed req. duration",
		},
		Rows: [][]string{
			{
				time.Since(f.started).String(),
				fmt.Sprintf("%d", crps),
				fmt.Sprintf("%d", atomic.LoadInt64(&f.actwks)),
				fmt.Sprintf("%d", total),
				fmt.Sprintf("%d", success),
				fmt.Sprintf("%d", failed),
				avgTotalDur,
				avgSuccessDur,
				avgFailedDur,
			},
		},
	}

	if isSummary {
		f.display.DrawSummary(t)
		return
	}

	f.display.Draw(t)
}

func (f *Flooder) spawnThread(wg *sync.WaitGroup, requestSendTicker *time.Ticker, crps int) {
	// calculating a current rps
	trps := int(float64(f.rps) * 0.95)
	// check the current rpc is under the target rpc and number of active workers is under the max workers
	if crps < trps && atomic.LoadInt64(&f.actwks) < f.maxwks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-f.ctx.Done():
					return
				case <-requestSendTicker.C:
					f.sendRequest()
				}
			}
		}()
		atomic.AddInt64(&f.actwks, 1)
	}
}

func (f *Flooder) sendRequest() {
	rand.Seed(time.Now().UnixNano())

	url := fmt.Sprintf(
		"https://seo-php-swoole.lux.kube.xbet.lan/api/v1/pagedata?group_id=285&u"+
			"rl=php-swoole-test-domain.com/fr&geo=by&language=fr&project[id]=285&domain=php-swoo"+
			"le-test-domain.com&timezone=3&stream=live&section=sport&sport[id]=1&timestamp=%d",
		rand.Uint64(),
	)

	s := time.Now()
	defer func() {
		atomic.AddInt64(&f.totalDurationMs, time.Since(s).Milliseconds())
	}()

	resp, err := http.Get(url)
	atomic.AddInt64(&f.total, 1)
	if err != nil || resp.StatusCode != 200 {
		atomic.AddInt64(&f.failed, 1)
		atomic.AddInt64(&f.failedDurationMs, time.Since(s).Milliseconds())
		log.Println(err, resp.StatusCode)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	_, _ = io.Copy(io.Discard, resp.Body)

	atomic.AddInt64(&f.success, 1)
	atomic.AddInt64(&f.successDurationMs, time.Since(s).Milliseconds())
}
