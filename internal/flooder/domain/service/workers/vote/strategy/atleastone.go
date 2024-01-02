package votestrategy

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

type AtLeastOneVoter struct {
	voters    []voter.Voter
	cfg       *config.Config
	collector statservice.Collector
}

func NewAtLeastOneVoter(
	voters []voter.Voter,
	cfg *config.Config,
	collector statservice.Collector,
) *AtLeastOneVoter {
	return &AtLeastOneVoter{
		voters:    voters,
		cfg:       cfg,
		collector: collector,
	}
}

func (s *AtLeastOneVoter) IsFor() bool {
	for _, voter := range s.voters {
		if voter(s.cfg, s.collector) {
			return true
		}
	}
	return false
}
