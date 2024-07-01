package worker

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
	wg  *sync.WaitGroup

	sender    *sender.Http
	logger    logservice.Logger
	collector statservice.Collector

	closeCh chan struct{}
}

func NewManagerService(
	ctx context.Context,
	sender *sender.Http,
	logger logservice.Logger,
	collector statservice.Collector,
) *ManagerService {
	return &ManagerService{
		ctx:       ctx,
		wg:        &sync.WaitGroup{},
		sender:    sender,
		logger:    logger,
		collector: collector,
		closeCh:   make(chan struct{}, 1),
	}
}

func (m *ManagerService) SpawnOne(ctx context.Context, sendTicker *time.Ticker) {
	m.wg.Add(1)
	go func() {
		defer func() {
			m.collector.RemoveWorker()
			m.wg.Done()
		}()
		m.collector.AddWorker()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.closeCh:
				return
			case <-sendTicker.C:
				// the request will be enriched in middleware
				m.sender.Send(new(http.Request).WithContext(ctx))
			}
		}
	}()
}

func (m *ManagerService) CloseOne() {
	select {
	case m.closeCh <- struct{}{}:
	default:
	}
}

func (m *ManagerService) CloseAll(cancel context.CancelFunc) {
	cancel()
	m.wg.Wait()
	close(m.closeCh)
	m.logger.Println("workers.Manager.CloseAll(): all spawned workers are closed")
}
