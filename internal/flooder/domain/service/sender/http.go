package sender

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/req"
	respmiddleware "github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/resp"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient"
	httpclientmiddleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"net/http"
)

type Http struct {
	cfg        *config.Config
	logger     logservice.Logger
	httpClient httpclient.Pooled
	collector  statservice.Collector
}

func NewHttp(
	cfg *config.Config,
	logger logservice.Logger,
	httpClient httpclient.Pooled,
	collector statservice.Collector,
) *Http {
	s := &Http{
		cfg:        cfg,
		logger:     logger,
		httpClient: httpClient,
		collector:  collector,
	}
	s.addMiddlewares()
	return s
}

func (s *Http) Send(req *http.Request) {
	_, _ = s.httpClient.Do(req)
}

func (s *Http) addMiddlewares() {
	s.httpClient.
		OnReq(s.requestMiddlewares()...).
		OnResp(s.responseMiddlewares()...)
}

func (s *Http) requestMiddlewares() []httpclientmiddleware.RequestMiddlewareFunc {
	mdw := []httpclientmiddleware.RequestMiddlewareFunc{
		reqmiddleware.NewInitRequestMiddleware(s.logger).InitRequest,
		reqmiddleware.NewRandUrlMiddleware(s.cfg.URLs, s.logger).AddRandUrl,
	}

	if s.cfg.AddTimestampToUrl {
		mdw = append(mdw, reqmiddleware.NewAddTimestampMiddleware().AddTimestamp)
	}

	if len(s.cfg.RequestHeaders) > 0 {
		mdw = append(mdw, reqmiddleware.NewAddHeadersMiddlewareMiddleware(s.cfg.RequestHeaders, s.logger).AddHeaders)
	}

	return mdw
}

func (s *Http) responseMiddlewares() []httpclientmiddleware.ResponseMiddlewareFunc {
	mdw := []httpclientmiddleware.ResponseMiddlewareFunc{
		respmiddleware.NewErrorMiddleware(s.logger).HandleError,
		respmiddleware.NewStatusCodeMiddleware(s.logger).HandleStatusCode,
	}

	if s.cfg.ExpectedResponseData != "" {
		mdw = append(mdw, respmiddleware.NewExpectedDataMiddleware(s.cfg, s.logger).CheckData)
	}

	mdw = append(mdw, respmiddleware.NewMetricsMiddleware(s.logger, s.collector).CollectMetrics)
	mdw = append(mdw, respmiddleware.NewCloseBodyMiddleware(s.logger).CloseResponseBody)

	return mdw
}
