package httpclient

import (
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	"net/http"
)

type Pooled interface {
	Do(req *http.Request) (*http.Response, error)
	OnReq(middlewares ...httpclientmiddleware.RequestMiddlewareFunc) Pooled
	OnResp(middlewares ...httpclientmiddleware.ResponseMiddlewareFunc) Pooled
	Busy() int64
	Total() int64
}
