package closevoter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

func ByAvgDuration(cfg *config.Config, collector *statservice.Collector) func(cfg *config.Config, collector *statservice.Collector) bool {
	return func(cfg *config.Config, collector *statservice.Collector) bool {
		return false
	}
}
