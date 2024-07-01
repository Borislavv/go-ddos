package orchestrator

import (
	"context"
	"sync"
)

type Orchestrator interface {
	Run(ctx context.Context, wg *sync.WaitGroup)
}
