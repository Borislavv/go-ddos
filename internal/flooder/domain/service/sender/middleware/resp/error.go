package respmiddleware

import (
	"encoding/json"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
)

type ErrorMiddleware struct {
	logger logservice.Logger
}

func NewErrorMiddleware(logger logservice.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{logger: logger}
}

func (m *ErrorMiddleware) HandleError(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *floodermodel.Response) *floodermodel.Response {
		if resp.Err() != nil {
			resp.SetFailed()

			b, e := json.MarshalIndent(floodermodel.Log{Error: resp.Err().Error()}, "", "  ")
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp)
			}
			m.logger.Println(string(b))
		}

		return next.Handle(resp)
	})
}
