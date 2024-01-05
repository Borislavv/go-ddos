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
	return &ByInterval{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ByInterval) Vote() (weight enum.Weight, sleep time.Duration) {
	if s.collector.Workers() < s.cfg.MaxWorkers {
		return enum.TotallyFor, time.Millisecond * 500
	} else {
		return enum.For, time.Millisecond * 1000
	}
}
