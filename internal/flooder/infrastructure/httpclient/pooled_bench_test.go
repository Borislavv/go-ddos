package httpclient

import (
	"context"
	"ddos/config"
	reqmiddleware "ddos/internal/flooder/domain/service/sender/middleware/req"
	"ddos/internal/flooder/domain/service/sender/middleware/resp"
	"ddos/internal/flooder/infrastructure/httpclient/config"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkPooled_Do(b *testing.B) {
	expectedResp := "fooBarBaz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(expectedResp)); err != nil {
			b.Fatal(err)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &httpclientconfig.Config{
		PoolInitSize: 10,
		PoolMaxSize:  1024,
	}

	client, cancelPool := NewPool(ctx, cfg, func() *http.Client {
		return &http.Client{Timeout: time.Minute}
	})
	defer cancelPool()

	collector := statservice.NewCollectorService(ctx, client, time.Minute*5, 1)

	logger, loggerClose := logservice.NewAsync(ctx, &config.Config{MaxWorkers: 2, MaxRPS: 10})
	defer loggerClose()
	mw := respmiddleware.NewMetricsMiddleware(logger, collector)

	client.
		OnReq(reqmiddleware.NewTimestampMiddleware().AddTimestamp).
		OnResp(mw.CollectMetrics)

	b.ResetTimer()
	b.StartTimer()
	b.SetParallelism(10)
	b.N = 100000
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				b.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return
			}

			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				b.Fatal(err)
			}

			if string(bytes) != expectedResp {
				b.Fatalf("expected response '%v', gotten '%v'", expectedResp, string(bytes))
			}
			_ = resp.Body.Close()
		}
	})
	b.StopTimer()

	b.Logf(
		"\n"+
			"total: %d, total duration: %d, avg duration ms: %v\n"+
			"success: %d, success duration: %d, avg success ms: %v\n"+
			"failed: %d, failed success: %d, avg failed ms: %v",
		collector.TotalRequests(), collector.TotalRequestsDuration(), collector.AvgTotalRequestsDuration().String(),
		collector.SuccessRequests(), collector.SuccessRequestsDuration(), collector.AvgSuccessRequestsDuration().String(),
		collector.FailedRequests(), collector.FailedRequestsDuration(), collector.AvgFailedRequestsDuration().String(),
	)
}
