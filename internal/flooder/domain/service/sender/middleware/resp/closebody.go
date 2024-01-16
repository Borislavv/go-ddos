package respmiddleware

import (
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	httpclientmodel "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/model"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
)

type CloseBodyMiddleware struct {
	logger logservice.Logger
}

func NewCloseBodyMiddleware(logger logservice.Logger) *CloseBodyMiddleware {
	return &CloseBodyMiddleware{logger: logger}
}

func (m *CloseBodyMiddleware) CloseResponseBody(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *httpclientmodel.Response) *httpclientmodel.Response {
		if !resp.IsFailed() && resp.Resp().Body != nil {
			if err := resp.Resp().Body.Close(); err != nil {
				m.logger.Println(err.Error())
			}
		}

		return next.Handle(resp)
	})
}
