package respmiddleware

import (
	"encoding/json"
	"fmt"
	floodermodel "github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
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
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		if err == nil && resp != nil && resp.StatusCode != http.StatusOK {
			var msg string
			if resp.StatusCode == http.StatusInternalServerError && resp.Body != nil {
				b, e := io.ReadAll(resp.Body)
				if e != nil {
					m.logger.Println(e.Error())
					return next.Handle(resp, err)
				}
				msg = string(b)
			} else {
				msg = fmt.Sprintf("request failed, received status code %d", resp.StatusCode)
			}

			b, e := json.MarshalIndent(floodermodel.Log{Error: msg}, "", "  ")
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err)
			}
			m.logger.Println(string(b))
		}

		return next.Handle(resp, err)
	})
}
