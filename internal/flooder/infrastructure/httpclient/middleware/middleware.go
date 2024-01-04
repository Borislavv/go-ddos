package httpclientmiddleware

import (
	"net/http"
)

type RequestModifier interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestModifierFunc func(req *http.Request) (*http.Response, error)

func (m RequestModifierFunc) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

type ResponseHandler interface {
	Handle(resp *http.Response, err error) (*http.Response, error)
}

type ResponseHandlerFunc func(resp *http.Response, err error) (*http.Response, error)

func (m ResponseHandlerFunc) Handle(resp *http.Response, err error) (*http.Response, error) {
	return m(resp, err)
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
