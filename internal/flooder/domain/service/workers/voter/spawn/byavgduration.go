package balancerspawnvoter

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"time"
)

type ByAvgDuration struct {
	cfg       *config.Config
	collector statservice.Collector
}

func NewByAvgDuration(cfg *config.Config, collector statservice.Collector) *ByAvgDuration {
	collector.SetLastSpawnByAvgDuration()

	return &ByAvgDuration{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByAvgDuration) Vote() (weight enum.Weight) {
	if time.Since(s.collector.LastSpawnByAvgDuration()) < s.cfg.SpawnIntervalValue {
		return enum.Check
	}

	defer s.collector.SetLastSpawnByAvgDuration()

	if s.collector.AvgSuccessRequestsDuration() <
		time.Duration(float64(s.cfg.TargetAvgSuccessRequestsDurationValue.Nanoseconds())*(1-s.cfg.ToleranceCoefficient)) {
		if s.collector.Workers() < s.cfg.MaxWorkers {
			return enum.AbsolutelyFor
		} else {
			return enum.SureFor
		}
	}

	return enum.Check
}
