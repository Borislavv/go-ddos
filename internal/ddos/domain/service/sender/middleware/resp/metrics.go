package respmiddleware

import (
	reqmiddleware "ddos/internal/ddos/domain/service/sender/middleware/req"
	middleware "ddos/internal/ddos/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"net/http"
	"strconv"
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
	return middleware.ResponseHandlerFunc(func(resp *http.Response, err error) (*http.Response, error) {
		var duration time.Duration

		defer func() {
			m.collector.AddTotalRequest()
			m.collector.AddTotalRequestsDuration(duration)
			if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
				m.collector.AddFailedRequest()
				m.collector.AddFailedRequestsDuration(duration)
			} else {
				m.collector.AddSuccessRequest()
				m.collector.AddSuccessRequestsDuration(duration)
			}
		}()

		if resp != nil && resp.Request != nil && resp.Request.URL != nil {
			copyValues := resp.Request.URL.Query()
			if !copyValues.Has(reqmiddleware.Timestamp) {
				m.logger.Println("timestamp doesn't exists in the request query, unable to determine request duration")
				return next.Handle(resp, err)
			}
			timestampStr := copyValues.Get(reqmiddleware.Timestamp)
			if timestampStr == "" {
				m.logger.Println("timestamp value is empty, unable to determine request duration")
				return next.Handle(resp, err)
			}
			timestamp, perr := strconv.ParseInt(timestampStr, 10, 64)
			if perr != nil {
				m.logger.Println("timestamp cannot by cast to int, unable to determine request duration")
				return next.Handle(resp, err)
			}
			duration = time.Since(time.UnixMilli(timestamp))
		} else {
			duration = time.Duration(0)
		}

		return next.Handle(resp, err)
	})
}
