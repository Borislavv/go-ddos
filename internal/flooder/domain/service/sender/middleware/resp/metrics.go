package respmiddleware

import (
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"net/http"
	"time"
)

type MetricsMiddleware struct {
	logger    logservice.Logger
	collector statservice.Collector
}

func NewMetricsMiddleware(
	logger logservice.Logger,
	collector statservice.Collector,
) *MetricsMiddleware {
	return &MetricsMiddleware{
		logger:    logger,
		collector: collector,
	}
}

func (m *MetricsMiddleware) CollectMetrics(next middleware.ResponseHandler) middleware.ResponseHandler {
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error, timestamp int64) (*http.Response, error, int64) {
		duration := time.Since(time.UnixMilli(timestamp))

		m.collector.AddTotalRequest()
		m.collector.AddTotalRequestsDuration(duration)
		if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
			m.collector.AddFailedRequest()
			m.collector.AddFailedRequestsDuration(duration)
		} else {
			m.collector.AddSuccessRequest()
			m.collector.AddSuccessRequestsDuration(duration)
		}

		return next.Handle(resp, err, timestamp)
	})
}
