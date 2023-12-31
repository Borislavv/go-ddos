package respmiddleware

import (
	middleware "ddos/internal/ddos/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	"net/http"
)

type CloseBodyMiddleware struct {
	logger *logservice.Logger
}

func NewCloseBodyMiddleware(logger *logservice.Logger) *CloseBodyMiddleware {
	return &CloseBodyMiddleware{logger: logger}
}

func (m *CloseBodyMiddleware) CloseResponseBody(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		if resp != nil && resp.Body != nil {
			if e := resp.Body.Close(); e != nil {
				m.logger.Println(err.Error())
			}
		}

		return next.Handle(resp, err)
	})
}
