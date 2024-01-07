package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type AtLeastOneVoter struct {
	voters []voter.Voter
}

func NewAtLeastOneVoter(voters []voter.Voter) *AtLeastOneVoter {
	return &AtLeastOneVoter{voters: voters}
}

func (s *AtLeastOneVoter) IsFor() bool {
	for _, v := range s.voters {
		isFor, _ := v.Vote()
		if isFor {
			return true
		}
	}
	return false
}
