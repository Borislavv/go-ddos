package spawnvoter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

func ByInterval(cfg *config.Config, collector *statservice.Collector) func(cfg *config.Config, collector *statservice.Collector) bool {
	return func(cfg *config.Config, collector *statservice.Collector) bool {
		return false
	}
}
