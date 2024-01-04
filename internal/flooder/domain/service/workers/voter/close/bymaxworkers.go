package balancerclosevoter

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

type ByMaxWorkers struct {
	cfg       *config.Config
	collector statservice.Collector
}

func NewByMaxWorkers(cfg *config.Config, collector statservice.Collector) *ByMaxWorkers {
	return &ByMaxWorkers{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByMaxWorkers) Vote() (weight enum.Weight) {
	if s.collector.Workers() <= s.cfg.MaxWorkers {
		return enum.Check
	}

	return enum.AbsolutelyFor
}
