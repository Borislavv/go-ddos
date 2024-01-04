package balancerspawnvoter

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

type ByMinWorkers struct {
	cfg       *config.Config
	collector statservice.Collector
}

func NewByMinWorkers(cfg *config.Config, collector statservice.Collector) *ByMinWorkers {
	return &ByMinWorkers{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByMinWorkers) Vote() (weight enum.Weight) {
	if s.collector.Workers() >= s.cfg.MinWorkers {
		return enum.Check
	}

	return enum.AbsolutelyFor
}
