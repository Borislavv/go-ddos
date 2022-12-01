package votestrategy

import (
	"ddos/config"
	"ddos/internal/ddos/domain/model"
	statservice "ddos/internal/stat/domain/service"
)

type AtLeastOneVoter struct {
	voters    []model.Voter
	cfg       *config.Config
	collector *statservice.Collector
}

func NewAtLeastOneVoter(
	voters []model.Voter,
	cfg *config.Config,
	collector *statservice.Collector,
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
