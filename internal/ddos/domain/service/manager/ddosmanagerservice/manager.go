package ddosmanagerservice

import (
	"context"
	"ddos/internal/ddos/domain/service/sender"
	logservice "ddos/internal/log/domain/service"
	statservice "ddos/internal/stat/domain/service"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	ctx context.Context

	sender    *sender.Sender
	logger    logservice.Logger
	collector statservice.Collector

	closeOneCh chan struct{}
	closeAllCh chan struct{}
}

func NewManager(
	ctx context.Context,
	sender *sender.Sender,
	logger logservice.Logger,
	collector statservice.Collector,
) *Manager {
	return &Manager{
		ctx:        ctx,
		sender:     sender,
		logger:     logger,
		collector:  collector,
		closeOneCh: make(chan struct{}, 1),
		closeAllCh: make(chan struct{}),
	}
}

func (m *Manager) Spawn(sendTicker *time.Ticker, wg *sync.WaitGroup) {
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

func (m *Manager) Close() {
	m.closeOneCh <- struct{}{}
}

func (m *Manager) CloseAll() {
	close(m.closeAllCh)
}
