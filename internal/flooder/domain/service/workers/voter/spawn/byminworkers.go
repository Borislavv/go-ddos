package balancerspawnvoter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

func ByMinWorkers() func(cfg *config.Config, collector statservice.Collector) bool {
	return func(cfg *config.Config, collector statservice.Collector) bool {
		return cfg.MaxWorkers > collector.Workers()
	}
}
