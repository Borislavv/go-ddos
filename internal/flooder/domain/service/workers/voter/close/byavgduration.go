package balancerclosevoter

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
	collector.SetLastCloseByAvgDuration()

	return &ByAvgDuration{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByAvgDuration) Vote() (isFor bool, weight enum.Weight) {
	if time.Since(s.collector.LastCloseByAvgDuration()) < s.cfg.SpawnIntervalValue {
		return false, enum.Check
	}

	defer s.collector.SetLastCloseByAvgDuration()
	return s.collector.AvgSuccessRequestsDuration() >
		time.Duration(float64(s.cfg.TargetAvgSuccessRequestsDurationValue.Nanoseconds())*(1+s.cfg.ToleranceCoefficient)) &&
		s.collector.Workers() > s.cfg.MinWorkers, enum.TotallyFor
}
