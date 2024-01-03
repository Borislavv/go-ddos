package httpclient

import (
	"context"
	"github.com/Borislavv/go-ddos/config"
	reqmiddleware "github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/req"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/resp"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/config"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
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

	logger := logservice.NewAsync(ctx, &config.Config{MinWorkers: 2, MaxRPS: 10})
	defer func() { _ = logger.Close() }()

	client := NewPool(ctx, cfg, func() *http.Client {
		return &http.Client{Timeout: time.Minute}
	})
	defer func() { _ = client.Close() }()

	collector := statservice.NewCollectorService(ctx, logger, client, time.Minute*5, 1)

	client.
		OnReq(reqmiddleware.NewTimestampMiddleware().AddTimestamp).
		OnResp(respmiddleware.NewMetricsMiddleware(logger, collector).CollectMetrics)

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
