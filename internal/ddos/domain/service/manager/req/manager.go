package req

import (
	"context"
	"ddos/internal/ddos/domain/service/sender"
	statservice "ddos/internal/stat/domain/service"
	"net/http"
	"sync"
	"time"
)

type Manager struct {
	ctx context.Context

	sender    *sender.Sender
	collector statservice.Collector

	closeOneCh chan struct{}
	closeAllCh chan struct{}
}

func NewManager(
	ctx context.Context,
	sender *sender.Sender,
	collector statservice.Collector,
) *Manager {
	return &Manager{
		ctx:        ctx,
		sender:     sender,
		collector:  collector,
		closeOneCh: make(chan struct{}, 1),
		closeAllCh: make(chan struct{}),
	}
}

func (m *Manager) Spawn(sendTicker *time.Ticker, wg *sync.WaitGroup) {
	wg.Add(1)
	m.collector.AddWorker()

	go func() {
		defer func() {
			m.collector.RemoveWorker()
			wg.Done()
		}()
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
