package displayservice

import "sync"

type Renderer interface {
	Run(wg *sync.WaitGroup)
	Write(p []byte) (n int, err error)
}
