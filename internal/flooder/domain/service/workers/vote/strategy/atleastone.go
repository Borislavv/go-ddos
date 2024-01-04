package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type AtLeastOneVoter struct {
	spawnVoters []voter.Voter
	closeVoters []voter.Voter
}

func NewAtLeastOneVoter(
	spawnVoters []voter.Voter,
	closeVoters []voter.Voter,
) *AtLeastOneVoter {
	return &AtLeastOneVoter{
		spawnVoters: spawnVoters,
		closeVoters: closeVoters,
	}
}

func (s *AtLeastOneVoter) For() enum.Action {
	var forSpawn enum.Weight
	for _, v := range s.spawnVoters {
		w := v.Vote()
		if forSpawn < w {
			forSpawn = w
		}
	}

	var forClose enum.Weight
	for _, v := range s.closeVoters {
		w := v.Vote()
		if forClose < w {
			forClose = w
		}
	}

	if forSpawn > forClose {
		return enum.Spawn
	} else if forClose > forSpawn {
		return enum.Close
	} else {
		return enum.Await
	}
}
