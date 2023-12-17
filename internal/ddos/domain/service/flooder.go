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

	// settings
	rpc     int
	workers int

	// dependencies
	display *display.Display

	// metrics
	total   int64     // number of total reqs.
	success int64     // number of success reqs.
	failed  int64     // number of failed reqs.
	actwks  int64     // number of active workers
	started time.Time // time of the ddos started
}

func NewFlooder(ctx context.Context, rpc int, workers int, display *display.Display) *Flooder {
	return &Flooder{
		mu:      &sync.RWMutex{},
		ctx:     ctx,
		rpc:     rpc,
		workers: workers,
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

		requestSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.rpc)*1.10))
		defer requestSendTicker.Stop()

		threadSpawnTicker := time.NewTicker(time.Millisecond * 50)
		defer threadSpawnTicker.Stop()

		for {
			crpc := int(float64(atomic.LoadInt64(&f.total)) / time.Since(f.started).Seconds())

			select {
			case <-f.ctx.Done():
				log.Println("flooder: received ctx.Done")

				f.sendSummary(crpc)

				return
			case <-threadSpawnTicker.C:
				trpc := int(float64(f.rpc) * 0.95)
				if crpc < trpc && f.actwks < 1000 {
					f.spawnThread(wg, requestSendTicker)

					atomic.AddInt64(&f.actwks, 1)
				}
			default:
				f.sendStat(crpc)

				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
}

func (f *Flooder) sendStat(crpc int) {
	f.display.Draw(
		&displaymodel.Table{
			Header: []string{
				"duration",
				"rpc",
				"workers",
				"total reqs.",
				"success reqs.",
				"failed reqs.",
			},
			Rows: [][]string{
				{
					time.Since(f.started).String(),
					fmt.Sprintf("%d", crpc),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.actwks)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.total)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.success)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.failed)),
				},
			},
		},
	)
}

func (f *Flooder) sendSummary(crpc int) {
	f.display.DrawSummary(
		&displaymodel.Table{
			Header: []string{
				"duration",
				"rpc",
				"workers",
				"total reqs.",
				"success reqs.",
				"failed reqs.",
			},
			Rows: [][]string{
				{
					time.Since(f.started).String(),
					fmt.Sprintf("%d", crpc),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.actwks)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.total)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.success)),
					fmt.Sprintf("%d", atomic.LoadInt64(&f.failed)),
				},
			},
		},
	)
}

func (f *Flooder) spawnThread(wg *sync.WaitGroup, requestSendTicker *time.Ticker) {
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
}

func (f *Flooder) sendRequest() {
	rand.Seed(time.Now().UnixNano())

	url := fmt.Sprintf(
		"https://seo-php-swoole.lux.kube.xbet.lan/api/v1/pagedata?group_id=285&u"+
			"rl=php-swoole-test-domain.com/fr&geo=by&language=fr&project[id]=285&domain=php-swoo"+
			"le-test-domain.com&timezone=3&stream=live&section=sport&sport[id]=1&timestamp=%d",
		rand.Uint64(),
	)

	resp, err := http.Get(url)
	atomic.AddInt64(&f.total, 1)
	if err != nil || resp.StatusCode != 200 {
		atomic.AddInt64(&f.failed, 1)
		log.Println(err, resp.StatusCode)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	atomic.AddInt64(&f.success, 1)

	_, _ = io.Copy(io.Discard, resp.Body)
}
