package httpclientmiddleware

import (
	"net/http"
)

type RequestModifier interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type RequestModifierFunc func(req *http.Request) (resp *http.Response, err error)

func (m RequestModifierFunc) Do(req *http.Request) (resp *http.Response, err error) {
	return m(req)
}

type ResponseHandler interface {
	Handle(resp *http.Response, err error, timestamp int64) (*http.Response, error, int64)
}

type ResponseHandlerFunc func(resp *http.Response, err error, timestamp int64) (*http.Response, error, int64)

func (m ResponseHandlerFunc) Handle(resp *http.Response, err error, timestamp int64) (*http.Response, error, int64) {
	return m(resp, err, timestamp)
}

type RequestMiddleware interface {
	Exec(next RequestModifier) RequestModifier
}

type RequestMiddlewareFunc func(next RequestModifier) RequestModifier

func (m RequestMiddlewareFunc) Exec(next RequestModifier) RequestModifier {
	return m(next)
}

type ResponseMiddleware interface {
	Exec(next ResponseHandler) ResponseHandler
}

type ResponseMiddlewareFunc func(next ResponseHandler) ResponseHandler

func (m ResponseMiddlewareFunc) Exec(next ResponseHandler) ResponseHandler {
	return m(next)
}
