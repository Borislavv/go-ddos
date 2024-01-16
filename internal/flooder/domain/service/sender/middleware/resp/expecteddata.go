package respmiddleware

import (
	"encoding/json"
	"errors"
	"github.com/Borislavv/go-ddos/config"
	floodermodel "github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"io"
	"reflect"
)

var (
	MismatchedDataWasDetectedError = errors.New("mismatched data was detected")
)

type ExpectedDataMiddleware struct {
	cfg    *config.Config
	logger logservice.Logger
}

func NewExpectedDataMiddleware(cfg *config.Config, logger logservice.Logger) *ExpectedDataMiddleware {
	return &ExpectedDataMiddleware{cfg: cfg, logger: logger}
}

func (m *ExpectedDataMiddleware) CheckData(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *floodermodel.Response) *floodermodel.Response {
		if !resp.IsFailed() && resp.Resp().Body != nil {
			b, e := io.ReadAll(resp.Resp().Body)
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp)
			}

			var responseData, expectedData interface{}
			if e = json.Unmarshal(b, &responseData); e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp)
			}
			if e = json.Unmarshal([]byte(m.cfg.ExpectedResponseData), &expectedData); e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp)
			}

			if !reflect.DeepEqual(responseData, expectedData) {
				resp.SetFailed()

				log := floodermodel.Log{
					Error:      "mismatch data found",
					StatusCode: resp.Resp().StatusCode,
					Data: floodermodel.Data{
						Expected: expectedData,
						Gotten:   responseData,
					},
				}

				if resp.Resp().Request != nil && resp.Resp().Request.URL != nil {
					log.URL = resp.Resp().Request.URL.String()
				}

				p, er := json.MarshalIndent(log, "", " ")
				if er != nil {
					m.logger.Println(e.Error())
					return next.Handle(resp)
				}

				m.logger.Println(string(p))
			}
		}

		return next.Handle(resp)
	})
}
