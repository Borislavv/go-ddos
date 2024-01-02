package balancerclosevoter

import (
	"github.com/Borislavv/go-ddos/config"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

func ByMaxWorkers() func(cfg *config.Config, collector statservice.Collector) bool {
	return func(cfg *config.Config, collector statservice.Collector) bool {
		return collector.Workers() > cfg.InitWorkersNum
	}
}
