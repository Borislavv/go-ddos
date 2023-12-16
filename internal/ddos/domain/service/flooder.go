package service

import (
	"context"
	"ddos/internal/display/app"
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
	mu      *sync.RWMutex
	ctx     context.Context
	rpc     int
	workers int

	// todo вынести
	displayCh chan string
	dataCh    chan *app.Data

	// todo вынести
	total   int64
	success int64
	failed  int64
}

func NewFlooder(ctx context.Context, rpc int, workers int, dataCh chan *app.Data) *Flooder {
	return &Flooder{
		mu:        &sync.RWMutex{},
		ctx:       ctx,
		rpc:       rpc,
		workers:   workers,
		displayCh: make(chan string, 1000),
		dataCh:    dataCh,
	}
}

func (f *Flooder) Run() {
	s := time.Now()

	t := time.NewTicker(time.Second / time.Duration(float64(f.rpc)*1.20))
	defer t.Stop()

	rand.Seed(time.Now().UnixNano())

	dwg := &sync.WaitGroup{}
	dwg.Add(1)
	go f.handleDisplayer(dwg)
	defer dwg.Wait()
	defer close(f.displayCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	// init default number of workers
	w := int64(1)
	wg.Add(f.workers)
	for ; w <= int64(f.workers); w++ {
		f.spawn(w, wg, s, t)
	}

	wg.Add(1)
	go func() {
		spawnTicker := time.NewTicker(time.Millisecond * 100)
		defer wg.Done()
		defer spawnTicker.Stop()
		//f.print("start spawning a new threads")
		for {
			select {
			case <-f.ctx.Done():
				//f.print("spawning new treads was stopped!")
				return
			case <-spawnTicker.C:
				rpc := int(float64(f.total) / time.Since(s).Seconds())
				if rpc < int(float64(f.rpc)*0.95) && w < 1000 {
					//f.print(fmt.Sprintf("current RPC: %d, target RPC: %d", rpc, f.rpc))
					wg.Add(1)
					f.spawn(atomic.AddInt64(&w, 1), wg, s, t)
				} else {
					//f.print(fmt.Sprintf("current RPC: %d, target RPC: %d", rpc, f.rpc))
				}
			default:
				rpc := int(float64(f.total) / time.Since(s).Seconds())
				f.dataCh <- &app.Data{
					CurrentDuration:        time.Since(s),
					TargetRPC:              f.rpc,
					CurrentRPC:             rpc,
					CurrentWorkers:         w,
					CurrentTotalRequests:   f.total,
					CurrentFailedRequests:  f.failed,
					CurrentSuccessRequests: f.success,
				}
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
}

func (f *Flooder) spawn(w int64, wg *sync.WaitGroup, s time.Time, t *time.Ticker) {
	go func() {
		defer wg.Done()
		//f.print(fmt.Sprintf("spawned %d worker", w))
		for {
			select {
			case <-f.ctx.Done():
				//f.mu.RLock()
				//f.print(
				//	fmt.Sprintf(
				//		"total: %d, success: %d, failed: %d, duration: %s",
				//		f.total, f.success, f.failed, time.Since(s).String(),
				//	),
				//)
				//f.mu.RUnlock()
				return
			case <-t.C:
				f.sendRequest()
			}
		}
	}()
}

func (f *Flooder) sendRequest() {
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
		return
	}
	defer func() { _ = resp.Body.Close() }()

	atomic.AddInt64(&f.success, 1)
	//data, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	f.print(err.Error())
	//	return
	//}

	//f.print(string(data))

	_, _ = io.Copy(io.Discard, resp.Body)
}

func (f *Flooder) print(msg string) {
	f.displayCh <- msg
}

func (f *Flooder) handleDisplayer(wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range f.displayCh {
		log.Println(msg)
	}
}
