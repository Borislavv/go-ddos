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
	collector.SetLastSpawnByRPS()

	return &ByRPS{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByRPS) Vote() (weight enum.Weight) {
	if time.Since(s.collector.LastSpawnByRPS()) < s.cfg.SpawnIntervalValue {
		return enum.Check
	}

	defer s.collector.SetLastSpawnByRPS()

	if s.collector.RPS() < int64(float64(s.cfg.TargetRPS)*(1-s.cfg.ToleranceCoefficient)) {
		if s.collector.Workers() < s.cfg.MaxWorkers {
			return enum.SureFor
		} else {
			return enum.For
		}
	}

	return enum.Check
}
