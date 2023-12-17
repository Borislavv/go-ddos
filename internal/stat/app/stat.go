package stat

import (
	"context"
	statservice "ddos/internal/stat/domain/service"
	"sync"
)

type Stat struct {
	ctx       context.Context
	collector *statservice.Collector
}

func New(ctx context.Context) *Stat {
	return &Stat{
		ctx: ctx,
	}
}

func (s *Stat) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

}
