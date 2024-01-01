package closevoter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

func ByMaxWorkers() func(cfg *config.Config, collector statservice.Collector) bool {
	return func(cfg *config.Config, collector statservice.Collector) bool {
		return collector.Workers() > cfg.MaxWorkers
	}
}
