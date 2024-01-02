package sender

import (
	"ddos/config"
	"ddos/internal/ddos/domain/service/sender/middleware/req"
	respmiddleware "ddos/internal/ddos/domain/service/sender/middleware/resp"
	"ddos/internal/ddos/infrastructure/httpclient"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"net/http"
)

type Sender struct {
	cfg        *config.Config
	logger     logservice.Logger
	httpClient *httpclient.Pool
	collector  statservice.Collector
}

func NewSender(
	cfg *config.Config,
	logger logservice.Logger,
	httpClient *httpclient.Pool,
	collector statservice.Collector,
) *Sender {
	s := &Sender{
		cfg:        cfg,
		logger:     logger,
		httpClient: httpClient,
		collector:  collector,
	}
	s.addMiddlewares()
	return s
}

func (s *Sender) Send(req *http.Request) {
	_, _ = s.httpClient.Do(req)
}

func (s *Sender) addMiddlewares() {
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
			respmiddleware.NewCloseBodyMiddleware(s.logger).CloseResponseBody,
		)
}
