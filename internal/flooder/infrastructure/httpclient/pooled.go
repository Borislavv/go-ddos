package httpclient

import (
	"context"
	config "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/config"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	"net/http"
	"sync/atomic"
)

type CancelFunc func()

type Pool struct {
	ctx     context.Context
	cfg     *config.Config
	pool    chan *http.Client
	creator func() *http.Client
	cancel  context.CancelFunc
	req     middleware.RequestModifier
	resp    middleware.ResponseHandler
	conns   int64
}

func NewPool(ctx context.Context, cfg *config.Config, creator func() *http.Client) *Pool {
	if cfg.PoolInitSize > cfg.PoolMaxSize {
		cfg.PoolInitSize = cfg.PoolMaxSize
	}

	ctx, cancel := context.WithCancel(ctx)

	p := &Pool{
		ctx:     ctx,
		cfg:     cfg,
		pool:    make(chan *http.Client, cfg.PoolMaxSize),
		creator: creator,
		cancel:  cancel,
	}

	for i := int64(0); i < cfg.PoolInitSize; i++ {
		p.pool <- creator()
	}

	p.conns = int64(cfg.PoolInitSize)

	p.req = middleware.RequestModifierFunc(
		func(req *http.Request) (*http.Response, error) {
			c := p.get()
			defer p.put(c)
			return c.Do(req)
		},
	)

	p.resp = middleware.ResponseHandlerFunc(
		func(resp *http.Response, err error) (*http.Response, error) {
			return resp, err
		},
	)

	return p
}

func (p *Pool) Do(req *http.Request) (*http.Response, error) {
	return p.resp.Handle(p.req.Do(req))
}

func (p *Pool) OnReq(middlewares ...middleware.RequestMiddlewareFunc) Pooled {
	for i := len(middlewares) - 1; i >= 0; i-- {
		p.req = middlewares[i].Exec(p.req)
	}
	return p
}

func (p *Pool) OnResp(middlewares ...middleware.ResponseMiddlewareFunc) Pooled {
	for i := len(middlewares) - 1; i >= 0; i-- {
		p.resp = middlewares[i].Exec(p.resp)
	}
	return p
}

func (p *Pool) Busy() int64 {
	return p.Total() - int64(len(p.pool))
}

func (p *Pool) Total() int64 {
	t := atomic.LoadInt64(&p.conns)
	if t > p.cfg.PoolMaxSize {
		return p.cfg.PoolMaxSize
	}
	return t
}

func (p *Pool) OutOfPool() int64 {
	return atomic.LoadInt64(&p.conns) - int64(cap(p.pool))
}

func (p *Pool) get() *http.Client {
	select {
	case c := <-p.pool:
		return c
	default:
		atomic.AddInt64(&p.conns, 1)
		return p.creator()
	}
}

func (p *Pool) put(c *http.Client) {
	select {
	case p.pool <- c:
	default:
	}
}

func (p *Pool) Close() error {
	p.cancel()
	close(p.pool)
	return nil
}
