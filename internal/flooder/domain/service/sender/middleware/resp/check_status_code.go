package respmiddleware

import (
	"encoding/json"
	"fmt"
	floodermodel "github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	httpclientmodel "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/model"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"io"
	"net/http"
)

type StatusCodeMiddleware struct {
	logger logservice.Logger
}

func NewStatusCodeMiddleware(logger logservice.Logger) *StatusCodeMiddleware {
	return &StatusCodeMiddleware{logger: logger}
}

func (m *StatusCodeMiddleware) HandleStatusCode(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *httpclientmodel.Response) *httpclientmodel.Response {
		if !resp.IsFailed() && resp.Resp() != nil && resp.Resp().StatusCode != http.StatusOK {
			resp.SetFailed()

			var msg string
			if resp.Resp().StatusCode == http.StatusInternalServerError && resp.Resp().Body != nil {
				b, e := io.ReadAll(resp.Resp().Body)
				if e != nil {
					m.logger.Println(e.Error())
					return next.Handle(resp)
				}
				msg = string(b)
			} else {
				msg = fmt.Sprintf("request failed, received status code %d", resp.Resp().StatusCode)
			}

			b, e := json.MarshalIndent(floodermodel.Log{Error: msg}, "", "  ")
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp)
			}
			m.logger.Println(string(b))
		}

		return next.Handle(resp)
	})
}
