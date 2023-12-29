package respmiddleware

import (
	"ddos/internal/ddos/domain/model"
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	"encoding/json"
	"net/http"
)

type ErrorMiddleware struct {
	logger *logservice.Logger
}

func NewErrorMiddleware(logger *logservice.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{logger: logger}
}

func (m *ErrorMiddleware) HandleError(next httpclientmiddleware.ResponseHandler) httpclientmiddleware.ResponseHandler {
	return httpclientmiddleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		if err != nil {
			b, e := json.MarshalIndent(model.Log{Error: err.Error()}, "", "  ")
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err)
			}
			m.logger.Println(string(b))
		}

		return next.Handle(resp, err)
	})
}
