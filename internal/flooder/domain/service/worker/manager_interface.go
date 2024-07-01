package worker

import (
	"context"
	"time"
)

type Manager interface {
	SpawnOne(ctx context.Context, sendTicker *time.Ticker)
	CloseOne()
	CloseAll(cancel context.CancelFunc)
}
