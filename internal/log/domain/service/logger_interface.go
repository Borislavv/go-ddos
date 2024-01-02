package logservice

import "sync"

type Logger interface {
	Run(wg *sync.WaitGroup)
	Println(msg string)
	Printf(msg string, args ...any)
}
