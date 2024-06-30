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
	rpsTrend  []int64
}

func NewByRPS(cfg *config.Config, collector statservice.Collector) *ByRPS {
	return &ByRPS{
		cfg:       cfg,
		collector: collector,
		rpsTrend:  make([]int64, 0, 100),
	}
}

func (s *ByRPS) Vote() (weight enum.Weight, sleep time.Duration) {
	currentRps := s.collector.RPS()

	if len(s.rpsTrend) == cap(s.rpsTrend) {
		s.rpsTrend = s.rpsTrend[10:]
	}
	s.rpsTrend = append(s.rpsTrend, currentRps)

	if currentRps > int64(float64(s.cfg.TargetRPS)*(1+s.cfg.ToleranceCoefficient)) {
		if s.collector.Workers() > s.cfg.MinWorkers {
			return enum.SureFor, s.cfg.SpawnIntervalValue * 2
		} else {
			return enum.Check, 0
		}
	} else {
		isFor := false
		if len(s.rpsTrend) == cap(s.rpsTrend) {
			isFor = s.rpsTrend[0] < s.rpsTrend[len(s.rpsTrend)-1]
		}
		if isFor {
			if s.collector.Workers() > s.cfg.MinWorkers {
				return enum.AbsolutelyFor, s.cfg.SpawnIntervalValue / 4
			}
		}
	}

	return enum.Check, s.cfg.SpawnIntervalValue * 2
}
