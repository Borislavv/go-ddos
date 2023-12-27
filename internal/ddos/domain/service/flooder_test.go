package ddosservice

import (
	"context"
	"ddos/config"
	display "ddos/internal/display/app"
	displayservice "ddos/internal/display/domain/service"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func BenchmarkFlooder_sendRequest(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("stress test of Flooder.Run")); err != nil {
			b.Log(err)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		MaxRPS:     10000,
		Stages:     1,
		MaxWorkers: 10,
		Duration:   "10m",
		URL:        fmt.Sprintf("%v?foo=bar", server.URL),
	}

	exitCh := make(chan os.Signal, 1)
	logsCh := make(chan string, int64(cfg.MaxRPS)*cfg.MaxWorkers)

	logger := logservice.NewLogger(ctx, cfg, logsCh)
	renderer := displayservice.NewRenderer(ctx, cfg, exitCh)
	displayer := display.New(ctx, renderer)
	collector := statservice.NewCollector(cfg)
	flooder := NewFlooder(ctx, cfg, displayer, logger, collector)

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go flooder.spawnResponseProcessor(wg)
	}
	defer close(flooder.respProcCh)

	b.ResetTimer()
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			flooder.sendRequest()
		}
	})
}
