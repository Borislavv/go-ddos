package balancerspawnvoter

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"time"
)

type ByRPS struct {
	cfg       *config.Config
	collector statservice.Collector
}

func NewByRPS(cfg *config.Config, collector statservice.Collector) *ByRPS {
	return &ByRPS{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByRPS) Vote() (weight enum.Weight, sleep time.Duration) {
	if s.collector.RPS() < int64(float64(s.cfg.TargetRPS)*(1-s.cfg.ToleranceCoefficient)) {
		if s.collector.Workers() < s.cfg.MaxWorkers {
			return enum.AbsolutelyFor, s.cfg.SpawnIntervalValue / 3
		} else {
			return enum.For, s.cfg.SpawnIntervalValue * 2
		}
	}

	return enum.Check, 0
}
