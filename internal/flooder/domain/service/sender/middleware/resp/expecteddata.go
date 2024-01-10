package respmiddleware

import (
	"encoding/json"
	"errors"
	"github.com/Borislavv/go-ddos/config"
	floodermodel "github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	"io"
	"net/http"
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
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		if resp != nil && resp.Body != nil && m.cfg.ExpectedResponseData != "" {
			b, e := io.ReadAll(resp.Body)
			if e != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err)
			}

			var responseData, expectedData interface{}
			if e = json.Unmarshal(b, &responseData); err != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err)
			}
			if e = json.Unmarshal([]byte(m.cfg.ExpectedResponseData), &expectedData); err != nil {
				m.logger.Println(e.Error())
				return next.Handle(resp, err)
			}

			if !reflect.DeepEqual(responseData, expectedData) {
				log := floodermodel.Log{
					Error:      "mismatch data found",
					StatusCode: resp.StatusCode,
					Data: floodermodel.Data{
						Expected: expectedData,
						Gotten:   responseData,
					},
				}

				if resp.Request != nil && resp.Request.URL != nil {
					log.URL = resp.Request.URL.String()
				}

				p, er := json.MarshalIndent(log, "", " ")
				if er != nil {
					m.logger.Println(e.Error())
					return next.Handle(resp, err)
				}

				m.logger.Println(MismatchedDataWasDetectedError.Error())

				if err == nil {
					return next.Handle(resp, MismatchedDataWasDetectedError)
				}
			}
		}

		return next.Handle(resp, err)
	})
}
