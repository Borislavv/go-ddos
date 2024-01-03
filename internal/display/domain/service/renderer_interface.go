package displayservice

import (
	displaymodel "ddos/internal/display/domain/model"
	"sync"
)

type Renderer interface {
	Draw(wg *sync.WaitGroup)
	Listen(wg *sync.WaitGroup)
	TableCh() chan<- *displaymodel.Table
	SummaryTableCh() chan<- *displaymodel.Table
	Close() error
}
