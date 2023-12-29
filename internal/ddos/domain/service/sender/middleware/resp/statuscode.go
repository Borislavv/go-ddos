package respmiddleware

import (
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	"net/http"
)

type StatusCodeMiddleware struct {
	logger *logservice.Logger
}

func NewStatusCodeMiddleware(logger *logservice.Logger) *StatusCodeMiddleware {
	return &StatusCodeMiddleware{logger: logger}
}

func (m *StatusCodeMiddleware) HandleStatusCode(next httpclientmiddleware.ResponseHandler) httpclientmiddleware.ResponseHandler {
	return httpclientmiddleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		if err != nil && resp.StatusCode != http.StatusOK {
			m.logger.Printf("request failed, received status code %d", resp.StatusCode)
		}

		return next.Handle(resp, err)
	})
}
