package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type ManyVoters struct {
	spawnVoters []voter.Voter
	closeVoters []voter.Voter
}

func NewManyVoters(
	spawnVoters []voter.Voter,
	closeVoters []voter.Voter,
) *ManyVoters {
	return &ManyVoters{
		spawnVoters: spawnVoters,
		closeVoters: closeVoters,
	}
}

func (s *ManyVoters) For() enum.Action {
	var spawnWeight enum.Weight
	for _, v := range s.spawnVoters {
		spawnWeight += v.Vote()
	}

	var closeWeight enum.Weight
	for _, v := range s.closeVoters {
		closeWeight += v.Vote()
	}

	if spawnWeight > closeWeight {
		return enum.Spawn
	} else if closeWeight > spawnWeight {
		return enum.Close
	} else {
		return enum.Await
	}
}
