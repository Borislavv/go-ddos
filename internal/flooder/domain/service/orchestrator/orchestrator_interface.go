package orchestrator

import (
	"context"
)

type Orchestrator interface {
	Run(ctx context.Context)
}
