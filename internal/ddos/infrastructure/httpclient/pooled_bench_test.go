package httpclient

import (
	"context"
	"ddos/config"
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

	cfg := &config.Config{
		URL:                   server.URL,
		HttpClientPoolMinSize: 32,
		HttpClientPoolMaxSize: 1024,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, cancelPool := NewPool(
		ctx,
		cfg.HttpClientPoolMinSize,
		cfg.HttpClientPoolMaxSize,
		func() *http.Client {
			return &http.Client{Timeout: time.Minute}
		},
	)
	defer cancelPool()

	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StartTimer()
	b.SetParallelism(10)
	b.N = 100000
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
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
}
