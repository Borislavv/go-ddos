package app

import (
	"context"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"sync"
	"time"
)

type Data struct {
	CurrentDuration        time.Duration
	TargetRPC              int
	CurrentRPC             int
	CurrentWorkers         int64
	CurrentTotalRequests   int64
	CurrentFailedRequests  int64
	CurrentSuccessRequests int64
}

type Display struct {
	ctx    context.Context
	dataCh chan *Data
}

func NewDisplay(ctx context.Context, dataCh chan *Data) *Display {
	return &Display{ctx: ctx, dataCh: dataCh}
}

func (d *Display) Run(mwg *sync.WaitGroup) {
	defer mwg.Done()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go d.draw(wg)
	wg.Wait()
}

func (d *Display) draw(wg *sync.WaitGroup) {
	defer wg.Done()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"duration", "rpc", "workers", "total reqs.", "success reqs.", "failed reqs."})

	for {
		select {
		case <-d.ctx.Done():
			termbox.Close()
			return
		case data := <-d.dataCh:
			if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				log.Fatalln(err)
			}

			// Очиста старых строк
			table.ClearRows()
			// Добавление данных в таблицу
			table.Append(
				[]string{
					data.CurrentDuration.String(),
					fmt.Sprintf("%d", data.CurrentRPC),
					fmt.Sprintf("%d", data.CurrentWorkers),
					fmt.Sprintf("%d", data.CurrentTotalRequests),
					fmt.Sprintf("%d", data.CurrentSuccessRequests),
					fmt.Sprintf("%d", data.CurrentFailedRequests),
				},
			)
			// Отрисовка таблицы
			table.Render()

			if err := termbox.Flush(); err != nil {
				log.Fatalln(err)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (d *Display) drawText(x, y int, text string) {
	for i, char := range text {
		termbox.SetCell(x+i, y, char, termbox.ColorDefault, termbox.ColorDefault)
	}
}
