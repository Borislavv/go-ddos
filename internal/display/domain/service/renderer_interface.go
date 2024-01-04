package displayservice

import (
	"context"
	displaymodel "ddos/internal/display/domain/model"
	"sync"
)

type Renderer interface {
	Draw(wg *sync.WaitGroup, ctx context.Context)
	Listen(wg *sync.WaitGroup, cancel context.CancelFunc)
	TableCh() chan<- *displaymodel.Table
	SummaryTableCh() chan<- *displaymodel.Table
	Close() error
}
