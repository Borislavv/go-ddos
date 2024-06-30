package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
	"time"
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

func (v *AllVoters) For() (action enum.Action, sleep time.Duration) {
	var forSpawn int
	var slpSpawn time.Duration
	for _, c := range v.spawnVoters {
		w, s := c.Vote()
		if w == enum.Check {
			continue
		}
		forSpawn++
		slpSpawn += s
	}
	if forSpawn < len(v.spawnVoters) {
		forSpawn = enum.Check
	}

	var forClose int
	var slpClose time.Duration
	for _, c := range v.closeVoters {
		w, s := c.Vote()
		if w == enum.Check {
			continue
		}
		forClose++
		slpClose += s
	}
	if forClose < len(v.closeVoters) {
		forClose = enum.Check
	}

	if forSpawn != enum.Check && forClose != enum.Check {
		return enum.Await, slpClose + slpSpawn
	} else if forSpawn != enum.Check {
		return enum.Spawn, slpSpawn
	} else {
		return enum.Close, slpClose
	}
}
