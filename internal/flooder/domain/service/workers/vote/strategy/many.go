package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type ManyVoters struct {
	voters []voter.Voter
}

func NewManyVoters(voters []voter.Voter) *ManyVoters {
	return &ManyVoters{voters: voters}
}

func (s *ManyVoters) IsFor() bool {
	var forNumber int
	var prosWeight enum.Weight
	var consWeight enum.Weight
	for _, v := range s.voters {
		isFor, weight := v.Vote()
		if isFor {
			forNumber++
			prosWeight += weight
		} else {
			consWeight += weight
		}
	}
	return prosWeight > consWeight || forNumber > len(s.voters)-forNumber
}
