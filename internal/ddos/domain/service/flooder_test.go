package ddosservice

import (
	"context"
	"ddos/config"
	display "ddos/internal/display/app"
	displaymodel "ddos/internal/display/domain/model"
	displayservice "ddos/internal/display/domain/service"
	statservice "ddos/internal/stat/domain/service"
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
		MaxRPS:      10000000,
		Percentiles: 1,
		MaxWorkers:  10,
		Duration:    "10m",
		URL:         server.URL,
	}

	exitCh := make(chan os.Signal, 1)
	dtCh := make(chan *displaymodel.Table, cfg.MaxRPS)
	smCh := make(chan *displaymodel.Table)

	renderer := displayservice.NewRenderer(ctx, dtCh, smCh, exitCh)
	displayer := display.New(ctx, renderer)
	collector := statservice.NewCollector(cfg)

	wg := &sync.WaitGroup{}
	collector.Consume(wg)
	defer func() {
		collector.Close()
		wg.Wait()
	}()

	flooder := NewFlooder(ctx, cfg, displayer, collector)

	b.ResetTimer()
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			flooder.sendRequest()
		}
	})
}
