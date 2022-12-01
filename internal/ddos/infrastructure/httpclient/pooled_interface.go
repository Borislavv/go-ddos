package httpclient

import (
	"ddos/internal/ddos/infrastructure/http/middleware"
	"net/http"
)

type Pooled interface {
	Do(req *http.Request) (*http.Response, error)
	OneReq(middlewares ...middleware.RequestMiddlewareFunc) Pooled
	OneResp(middlewares ...middleware.ResponseMiddlewareFunc) Pooled
}
