package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
)

type AllVoters struct {
	spawnVoters []voter.Voter
	closeVoters []voter.Voter
}

func NewAllVoters(
	spawnVoters []voter.Voter,
	closeVoters []voter.Voter,
) *AllVoters {
	return &AllVoters{
		spawnVoters: spawnVoters,
		closeVoters: closeVoters,
	}
}

func (s *AllVoters) For() enum.Action {
	var forSpawn int
	for _, v := range s.spawnVoters {
		if v.Vote() == enum.Check {
			continue
		}
		forSpawn++
	}
	if forSpawn < len(s.spawnVoters) {
		forSpawn = enum.Check
	}

	var forClose int
	for _, v := range s.closeVoters {
		if v.Vote() == enum.Check {
			continue
		}
		forClose++
	}
	if forClose < len(s.spawnVoters) {
		forClose = enum.Check
	}

	if forSpawn != enum.Check && forClose != enum.Check {
		return enum.Await
	} else if forSpawn != enum.Check {
		return enum.Spawn
	} else {
		return enum.Close
	}
}
