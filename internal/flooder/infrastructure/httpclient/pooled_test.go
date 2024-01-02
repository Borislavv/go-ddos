package httpclient

import (
	"context"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/config"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

	cfg := &httpclientconfig.Config{
		PoolInitSize: 10,
		PoolMaxSize:  1024,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := NewPool(ctx, cfg, func() *http.Client {
		return &http.Client{Timeout: time.Minute}
	})
	defer func() { _ = client.Close() }()

	req, err := http.NewRequest("GET", server.URL, nil)
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

	cfg := &httpclientconfig.Config{
		PoolInitSize: 10,
		PoolMaxSize:  1024,
	}

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := NewPool(context.Background(), cfg, func() *http.Client {
		return &http.Client{Timeout: time.Minute}
	})
	defer func() { _ = client.Close() }()

	type Counter struct {
		Responses int
	}
	counter := &Counter{}

	client.
		OnReq(
			func(next httpclientmiddleware.RequestModifier) httpclientmiddleware.RequestModifier {
				return httpclientmiddleware.RequestModifierFunc(func(req *http.Request) (*http.Response, error) {
					copyValues := req.URL.Query()
					if copyValues.Has("timestamp") {
						copyValues.Del("timestamp")
					}
					copyValues.Add("timestamp", strconv.Itoa(time.Now().Nanosecond()))
					req.URL.RawQuery = copyValues.Encode()
					return next.Do(req)
				})
			},
		).
		OnResp(
			func(next httpclientmiddleware.ResponseHandler) httpclientmiddleware.ResponseHandler {
				return httpclientmiddleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
					counter.Responses++
					return next.Handle(resp, err)
				})
			},
			func(next httpclientmiddleware.ResponseHandler) httpclientmiddleware.ResponseHandler {
				return httpclientmiddleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
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
}
