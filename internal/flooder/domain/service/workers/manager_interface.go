package workers

import (
	"sync"
	"time"
)

type Manager interface {
	Spawn(wg *sync.WaitGroup, sendTicker *time.Ticker)
	Close()
	CloseAll()
}
