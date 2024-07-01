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
	return &ByAvgDuration{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByAvgDuration) Vote() (weight enum.Weight, sleep time.Duration) {
	if s.collector.Workers() >= s.cfg.MaxWorkers {
		return enum.Check, 0
	}

	mn := time.Duration(float64(s.cfg.TargetAvgSuccessRequestsDurationValue) * (1 - s.cfg.ToleranceCoefficient))
	cr := s.collector.AvgSuccessRequestsDuration()

	if cr > mn || cr == 0 {
		return enum.Check, 0
	}

	cf := float64(mn) / float64(cr)
	if cf <= 0 {
		return enum.Check, 0
	}

	if cf >= 2 {
		return enum.TotallyFor, time.Millisecond * 50
	}
	if cf >= 1 {
		return enum.TotallyFor, time.Millisecond * 250
	}
	if cf >= 0.5 {
		return enum.SureFor, time.Millisecond * 750
	}
	if cf >= 0.2 {
		return enum.For, time.Millisecond * 1500
	}
	if cf >= 0.1 {
		return enum.For, time.Millisecond * 3000
	}

	return enum.Check, 0
}
