package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type AllVoters struct {
	voters []voter.Voter
}

func NewAllVoters(voters []voter.Voter) *AllVoters {
	return &AllVoters{voters: voters}
}

func (s *AllVoters) IsFor() bool {
	for _, v := range s.voters {
		isFor, _ := v.Vote()
		if !isFor {
			return false
		}
	}
	return true
}
