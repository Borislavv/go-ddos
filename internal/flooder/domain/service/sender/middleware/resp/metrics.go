package respmiddleware

import (
	floodermodel "github.com/Borislavv/go-ddos/internal/flooder/domain/model"
	middleware "github.com/Borislavv/go-ddos/internal/flooder/infrastructure/httpclient/middleware"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
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
	return middleware.ResponseHandlerFunc(func(resp *floodermodel.Response) *floodermodel.Response {
		duration := resp.Duration()

		m.collector.AddTotalRequest()
		m.collector.AddTotalRequestsDuration(duration)

		if resp.IsFailed() {
			m.collector.AddFailedRequest()
			m.collector.AddFailedRequestsDuration(duration)
		} else {
			m.collector.AddSuccessRequest()
			m.collector.AddSuccessRequestsDuration(duration)
		}

		return next.Handle(resp)
	})
}
