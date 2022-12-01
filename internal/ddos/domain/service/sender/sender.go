package req

import (
	"ddos/internal/ddos/domain/service/processor/resp"
	reqmiddleware "ddos/internal/ddos/domain/service/sender/req/middleware"
	"ddos/internal/ddos/infrastructure/httpclient"
	"ddos/internal/ddos/infrastructure/httpclient/middleware"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"fmt"
	"net/http"
	"time"
)

type Sender struct {
	processor       *resp.Processor
	logger          *logservice.Logger
	httpClient      *httpclient.Pool
	collector       *statservice.Collector
	reqMiddlewares  []httpclientmiddleware.RequestModifier
	respMiddlewares []httpclientmiddleware.ResponseHandler
}

func NewSender(
	processor *resp.Processor,
	logger *logservice.Logger,
	httpClient *httpclient.Pool,
	collector *statservice.Collector,
	reqMiddlewares []httpclientmiddleware.RequestModifier,
	respMiddlewares []httpclientmiddleware.ResponseHandler,
) *Sender {
	return &Sender{
		processor:       processor,
		logger:          logger,
		httpClient:      httpClient,
		collector:       collector,
		reqMiddlewares:  reqMiddlewares,
		respMiddlewares: respMiddlewares,
	}
}

func (s *Sender) Send(req *http.Request) {
	st := time.Now()
	defer func() {
		s.collector.AddTotal()
		s.collector.AddTotalDuration(time.Since(st))
	}()

	response, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Println(err.Error())
		s.collector.AddFailed()
		s.collector.AddFailedDuration(time.Since(st))
		return
	} else if response.StatusCode != http.StatusOK {
		defer func() { _ = response.Body.Close() }()
		s.logger.Println(fmt.Sprintf("Status code: %d", response.StatusCode))
		s.collector.AddFailed()
		s.collector.AddFailedDuration(time.Since(st))
		return
	} else {
		defer func() { _ = response.Body.Close() }()
		s.collector.AddSuccess()
		s.collector.AddSuccessDuration(time.Since(st))
		return
	}
}

func (s *Sender) addRequestMiddlewares() {
	s.httpClient.
		OnReq(
			reqmiddleware.RandUrl,
			reqmiddleware.AddTimestamp,
		)
}

func (s *Sender) addResponseMiddlewares() {
	s.httpClient.
		OnReq(
			reqmiddleware.AddTimestamp,
		)
}
