package httpclient

import (
	"context"
	"ddos/config"
	reqmiddleware "ddos/internal/ddos/domain/service/sender/req/middleware"
	"ddos/internal/ddos/infrastructure/http/middleware"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPooled_Do(t *testing.T) {
	expectedResp := "fooBarBaz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(expectedResp)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		URL:                   server.URL,
		HttpClientPoolMinSize: 10,
		HttpClientPoolMaxSize: 100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, clientCancel := NewPool(
		ctx,
		cfg.HttpClientPoolMinSize,
		cfg.HttpClientPoolMaxSize,
		func() *http.Client {
			return &http.Client{Timeout: time.Minute}
		},
	)
	defer clientCancel()

	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != expectedResp {
		t.Fatalf("expected response '%v', gotten '%v'", expectedResp, string(b))
	}
}

func TestPooled_Use(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("")); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		URL:                   server.URL,
		HttpClientPoolMinSize: 10,
		HttpClientPoolMaxSize: 100,
	}

	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client, clientCancel := NewPool(
		context.Background(),
		cfg.HttpClientPoolMinSize,
		cfg.HttpClientPoolMaxSize,
		func() *http.Client {
			return &http.Client{Timeout: time.Minute}
		},
	)
	defer clientCancel()

	type Counter struct {
		Responses int
	}
	counter := &Counter{}

	client.
		OneReq(
			reqmiddleware.AddTimestamp,
		).
		OneResp(
			func(next middleware.ResponseHandler) middleware.ResponseHandler {
				return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
					counter.Responses++
					return next.Handle(resp, err)
				})
			},
			func(next middleware.ResponseHandler) middleware.ResponseHandler {
				return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
					counter.Responses++
					return next.Handle(resp, err)
				})
			},
		)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.Request.URL.Query().Get("timestamp") == "" {
		t.Fatal("request middleware does not applied, timestamp does not exists into the url")
	}

	if counter.Responses != 4 {
		t.Fatal("response middleware does not applied, counter is equals zero")
	}

	time.Sleep(time.Second)
}
