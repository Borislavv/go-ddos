package voter

import (
	"ddos/config"
	statservice "ddos/internal/stat/domain/service"
)

type Voter func(cfg *config.Config, collector statservice.Collector) bool
