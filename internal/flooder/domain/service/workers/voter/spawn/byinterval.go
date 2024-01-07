package balancerspawnvoter

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
	"time"
)

type ByInterval struct {
	cfg       *config.Config
	collector statservice.Collector
}

func NewByInterval(cfg *config.Config, collector statservice.Collector) *ByInterval {
	collector.SetLastSpawnByInterval()

	return &ByInterval{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByInterval) Vote() (isFor bool, weight enum.Weight) {
	defer s.collector.SetLastSpawnByInterval()
	return time.Since(s.collector.LastSpawnByInterval()) > s.cfg.SpawnIntervalValue &&
		s.collector.Workers() < s.cfg.MaxWorkers, enum.TotallyFor
}
