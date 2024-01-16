package httpclientmiddleware

import (
	model "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/model"
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
	Handle(resp *model.Response) *model.Response
}

type ResponseHandlerFunc func(resp *model.Response) *model.Response

func (m ResponseHandlerFunc) Handle(resp *model.Response) *model.Response {
	return m(resp)
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
