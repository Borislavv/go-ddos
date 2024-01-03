package displayservice

import (
	"context"
	"ddos/config"
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
	isClosed       int64
}

func NewRenderer(
	ctx context.Context,
	cfg *config.Config,
	logger logservice.Logger,
	exitCh chan<- os.Signal,
) *RendererService {
	return &RendererService{
		ctx:            ctx,
		exitCh:         exitCh,
		logger:         logger,
		tableCh:        make(chan *displaymodel.Table, cfg.MaxWorkers),
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

func (r *RendererService) Listen(wg *sync.WaitGroup) {
	defer func() {
		close(r.stopCh)
		wg.Done()
	}()

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			switch event := termbox.PollEvent(); event.Type {
			case termbox.EventKey:
				if event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlZ {
					return
				}
			case termbox.EventError:
				r.logger.Println(event.Err.Error())
			}
		}
	}
}

func (r *RendererService) Draw(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := termbox.Init(); err != nil {
		r.logger.Println(err.Error())
		return
	}
	defer termbox.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	for {
		select {
		case <-r.stopCh:
			termbox.Close()

			r.exitCh <- os.Interrupt

			data := <-r.summaryTableCh

			// reinitialized the table as the summary
			table = tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_CENTER)

			// draw the summary table
			if err := r.draw(table, data); err != nil {
				r.logger.Println(err.Error())
				continue
			}

			return
		case data := <-r.tableCh:
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
