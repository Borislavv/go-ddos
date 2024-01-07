package balancerclosevoter

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
	collector.SetLastCloseByRPS()

	return &ByRPS{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByRPS) Vote() (isFor bool, weight enum.Weight) {
	if time.Since(s.collector.LastCloseByRPS()) < s.cfg.SpawnIntervalValue {
		return false, enum.Check
	}

	defer s.collector.SetLastCloseByRPS()
	return s.collector.RPS() > int64(float64(s.cfg.TargetRPS)*(1+s.cfg.ToleranceCoefficient)) &&
		s.collector.Workers() > s.cfg.MinWorkers, enum.For
}
