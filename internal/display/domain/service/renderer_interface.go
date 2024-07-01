package displayservice

import (
	"context"
	"sync"
)

type Renderer interface {
	Run(ctx context.Context, wg *sync.WaitGroup)
	Write(p []byte) (n int, err error)
}
