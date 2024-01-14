package respmiddleware

import (
	"encoding/json"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"net/http"
)

type ErrorMiddleware struct {
	logger logservice.Logger
}

func NewErrorMiddleware(logger logservice.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{logger: logger}
}

func (m *ErrorMiddleware) HandleError(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error, timestamp int64) (*http.Response, error, int64) {
		if err != nil {
			b, e := json.MarshalIndent(floodermodel.Log{Error: err.Error()}, "", "  ")
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err, timestamp)
			}
			m.logger.Println(string(b))
		}

		return next.Handle(resp, err, timestamp)
	})
}
