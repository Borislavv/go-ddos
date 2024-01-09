package sender

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/req"
	respmiddleware "github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender/middleware/resp"
	"github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient"
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
		OnReq(
			reqmiddleware.NewInitRequestMiddleware(s.logger).InitRequest,
			reqmiddleware.NewRandUrlMiddleware([]string{s.cfg.URL}, s.logger).AddRandUrl,
			reqmiddleware.NewTimestampMiddleware().AddTimestamp,
		).
		OnResp(
			respmiddleware.NewErrorMiddleware(s.logger).HandleError,
			respmiddleware.NewStatusCodeMiddleware(s.logger).HandleStatusCode,
			respmiddleware.NewMetricsMiddleware(s.logger, s.collector).CollectMetrics,
			respmiddleware.NewExpectedDataMiddleware(s.cfg, s.logger).CheckData,
			respmiddleware.NewCloseBodyMiddleware(s.logger).CloseResponseBody,
		)
}
