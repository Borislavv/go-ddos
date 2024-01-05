package votestrategy

import (
	"ddos/config"
	"ddos/internal/flooder/domain/service/workers/voter"
	statservice "ddos/internal/stat/domain/service"
)

type AllVoters struct {
	voters    []voter.Voter
	cfg       *config.Config
	collector statservice.Collector
}

func NewAllVoters(
	voters []voter.Voter,
	cfg *config.Config,
	collector statservice.Collector,
) *AllVoters {
	return &AllVoters{
		voters:    voters,
		cfg:       cfg,
		collector: collector,
	}
}

func (s *AllVoters) IsFor() bool {
	for _, voter := range s.voters {
		if !voter(s.cfg, s.collector) {
			return false
		}
	}
	return true
}
