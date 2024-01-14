package httpclient

import (
	httpclientmiddleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	"net/http"
)

type Pooled interface {
	Do(req *http.Request) (resp *http.Response, err error, timestamp int64)
	OnReq(middlewares ...httpclientmiddleware.RequestMiddlewareFunc) Pooled
	OnResp(middlewares ...httpclientmiddleware.ResponseMiddlewareFunc) Pooled

	Busy() int64
	Total() int64
	OutOfPool() int64
}
