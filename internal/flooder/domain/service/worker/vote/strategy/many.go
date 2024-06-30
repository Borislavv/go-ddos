package votestrategy

import (
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/worker/voter"
	"time"
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

func (v *ManyVoters) For() (action enum.Action, sleep time.Duration) {
	var forSpawn enum.Weight
	var slpSpawn time.Duration
	for _, c := range v.spawnVoters {
		w, s := c.Vote()
		forSpawn += w
		slpSpawn += s
	}

	var forClose enum.Weight
	var slpClose time.Duration
	for _, c := range v.closeVoters {
		w, s := c.Vote()
		forClose += w
		slpClose += s
	}

	if forSpawn > forClose {
		return enum.Spawn, slpSpawn
	} else if forClose > forSpawn {
		return enum.Close, slpClose
	} else {
		return enum.Await, slpSpawn + slpClose
	}
}
