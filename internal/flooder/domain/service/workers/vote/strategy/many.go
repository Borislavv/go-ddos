package votestrategy

import (
	"ddos/config"
	"ddos/internal/flooder/domain/service/balancer"
	statservice "ddos/internal/stat/domain/service"
)

type ManyVoters struct {
	voters    []balancer.Voter
	cfg       *config.Config
	collector statservice.Collector
}

func NewManyVoters(
	voters []balancer.Voter,
	cfg *config.Config,
	collector statservice.Collector,
) *ManyVoters {
	return &ManyVoters{
		voters:    voters,
		cfg:       cfg,
		collector: collector,
	}
}

func (s *ManyVoters) IsFor() bool {
	i := 0
	for _, voter := range s.voters {
		if voter(s.cfg, s.collector) {
			i++
		}
	}
	return i > len(s.voters)-i
}
