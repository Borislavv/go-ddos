package voter

import (
	"github.com/Borislavv/go-ddos/config"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

type Voter func(cfg *config.Config, collector statservice.Collector) bool
