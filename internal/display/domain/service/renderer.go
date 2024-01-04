package displayservice

import (
	"context"
	displaymodel "ddos/internal/display/domain/model"
	logservice "ddos/internal/log/domain/service"
	"github.com/nsf/termbox-go"
	"github.com/olekukonko/tablewriter"
	"os"
	"sync"
)

type RendererService struct {
	ctx            context.Context
	logger         logservice.Logger
	tableCh        chan *displaymodel.Table
	summaryTableCh chan *displaymodel.Table
	exitCh         chan<- os.Signal
	stopCh         chan struct{}
}

func NewRendererService(
	ctx context.Context,
	logger logservice.Logger,
	exitCh chan<- os.Signal,
) *RendererService {
	return &RendererService{
		ctx:            ctx,
		exitCh:         exitCh,
		logger:         logger,
		tableCh:        make(chan *displaymodel.Table, 1),
		summaryTableCh: make(chan *displaymodel.Table, 1),
		stopCh:         make(chan struct{}, 1),
	}
}

func (r *RendererService) TableCh() chan<- *displaymodel.Table {
	return r.tableCh
}

func (r *RendererService) SummaryTableCh() chan<- *displaymodel.Table {
	return r.summaryTableCh
}

func (r *RendererService) Listen(wg *sync.WaitGroup, cancel context.CancelFunc) {
	defer func() {
		r.logger.Println("display.RendererService.Listen() is closed")
		wg.Done()
	}()

	for {
		switch event := termbox.PollEvent(); event.Type {
		case termbox.EventKey:
			if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlZ {
				cancel()
				continue
			}
		case termbox.EventError:
			r.logger.Println("termbox receive error Event: " + event.Err.Error())
			continue
		case termbox.EventInterrupt:
			return
		}
	}
}

func (r *RendererService) Draw(wg *sync.WaitGroup, ctx context.Context) {
	defer func() {
		termbox.Interrupt()
		r.logger.Println("display.RendererService.Draw() is closed")
		wg.Done()
	}()

	if err := termbox.Init(); !termbox.IsInit && err != nil {
		r.logger.Println(err.Error())
		return
	}
	defer termbox.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	for {
		select {
		case <-ctx.Done():
			termbox.Close()

			r.exitCh <- os.Interrupt

			data, ok := <-r.summaryTableCh
			if !ok {
				return
			}

			// reinitialized the table as the summary
			table = tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_CENTER)

			// draw the summary table
			if err := r.draw(table, data); err != nil {
				r.logger.Println(err.Error())
				continue
			}
			return
		case data, isOpen := <-r.tableCh:
			if !isOpen {
				continue
			}

			// clear a terminal window
			if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
				r.logger.Println(err.Error())
				continue
			}
			// sync a terminal window
			if err := termbox.Sync(); err != nil {
				r.logger.Println(err.Error())
				continue
			}

			// draw a table
			if err := r.draw(table, data); err != nil {
				r.logger.Println(err.Error())
				continue
			}
		}
	}
}

func (r *RendererService) Close() error {
	close(r.tableCh)
	close(r.summaryTableCh)
	return nil
}

func (r *RendererService) draw(table *tablewriter.Table, data *displaymodel.Table) error {
	// set up a table header
	table.SetHeader(data.Header)

	// clear a table rows
	table.ClearRows()

	// set up a table rows
	table.AppendBulk(data.Rows)

	if data.Footer != nil {
		table.SetFooter(data.Footer)
	}

	// render a table
	table.Render()

	// draw a table
	if err := termbox.Flush(); err != nil {
		return err
	}

	return nil
}
