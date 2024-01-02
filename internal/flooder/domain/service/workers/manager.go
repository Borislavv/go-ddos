package workers

import (
	"context"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/sender"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"net/http"
	"sync"
	"time"
)

type ManagerService struct {
	ctx context.Context

	sender    *sender.Http
	logger    logservice.Logger
	collector statservice.Collector

	closeOneCh chan struct{}
	closeAllCh chan struct{}
}

func NewManagerService(
	ctx context.Context,
	sender *sender.Http,
	logger logservice.Logger,
	collector statservice.Collector,
) *ManagerService {
	return &ManagerService{
		ctx:        ctx,
		sender:     sender,
		logger:     logger,
		collector:  collector,
		closeOneCh: make(chan struct{}, 1),
		closeAllCh: make(chan struct{}),
	}
}

func (m *ManagerService) Spawn(wg *sync.WaitGroup, sendTicker *time.Ticker) {
	wg.Add(1)
	go func() {
		defer func() {
			m.collector.RemoveWorker()
			wg.Done()
		}()
		m.collector.AddWorker()
		for {
			select {
			case <-m.closeOneCh:
				return
			case <-m.closeAllCh:
				return
			case <-sendTicker.C:
				// sending a request, with a nil request struct specified
				// due to all useful work will be done in the middlewares
				m.sender.Send(new(http.Request))
			}
		}
	}()
}

func (m *ManagerService) Close() {
	m.closeOneCh <- struct{}{}
}

func (m *ManagerService) CloseAll() {
	close(m.closeAllCh)
}
