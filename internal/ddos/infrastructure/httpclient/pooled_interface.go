package httpclient

import (
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	"net/http"
)

type Pooled interface {
	Do(req *http.Request) (*http.Response, error)
	OnReq(middlewares ...httpclientmiddleware.RequestMiddlewareFunc) Pooled
	OnResp(middlewares ...httpclientmiddleware.ResponseMiddlewareFunc) Pooled
}
