package votestrategy

import (
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

type ManyVoters struct {
	voters    []voter.Voter
	cfg       *config.Config
	collector statservice.Collector
}

func NewManyVoters(
	voters []voter.Voter,
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
