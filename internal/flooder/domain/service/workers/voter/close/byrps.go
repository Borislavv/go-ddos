package balancerclosevoter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

func ByRPS() func(cfg *config.Config, collector statservice.Collector) bool {
	return func(cfg *config.Config, collector statservice.Collector) bool {
		return false
	}
}
