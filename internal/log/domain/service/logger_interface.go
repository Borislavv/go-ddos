package logservice

import "sync"

type Logger interface {
	Run(wg *sync.WaitGroup)
	Println(msg string)
	Printfln(msg string, args ...any)
}
