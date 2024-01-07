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

func (s *ByMinWorkers) Vote() (isFor bool, weight enum.Weight) {
	return s.cfg.MinWorkers > s.collector.Workers(), enum.TotallyFor
}
