package reqsender

import (
	"context"
	"ddos/config"
	"ddos/internal/ddos/domain/enum"
	"ddos/internal/ddos/domain/model"
	closevoter "ddos/internal/ddos/domain/service/balancer/request/voter/close"
	spawnvoter "ddos/internal/ddos/domain/service/balancer/request/voter/spawn"
	votestrategy "ddos/internal/ddos/domain/service/balancer/vote/strategy"
	statservice "ddos/internal/stat/domain/service"
	"errors"
)

var (
	SpawnVoteStrategyWasNotFoundError = errors.New("spawn vote strategy not found")
	CloseVoteStrategyWasNotFoundError = errors.New("close vote strategy not found")
)

type Balancer struct {
	ctx                  context.Context
	cfg                  *config.Config
	collector            *statservice.Collector
	voteStrategyForSpawn votestrategy.VoteStrategy
	voteStrategyForClose votestrategy.VoteStrategy
	votersForSpawn       []model.Voter
	votersForClose       []model.Voter
}

func NewBalancer(
	ctx context.Context,
	cfg *config.Config,
	collector *statservice.Collector,
) (*Balancer, error) {
	s := &Balancer{
		ctx:       ctx,
		cfg:       cfg,
		collector: collector,
	}

	if err := s.initVoteStrategyForSpawn(); err != nil {
		return nil, err
	}

	if err := s.initVoteStrategyForClose(); err != nil {
		return nil, err
	}

	s.initVotersForSpawn()
	s.initVotersForClose()

	return s, nil
}

func (s *Balancer) initVoteStrategyForSpawn() error {
	switch enum.VoteStrategy(s.cfg.VoteForSpawnReqSenderStrategy) {
	case enum.AllVotersStrategy:
		s.voteStrategyForSpawn = votestrategy.NewAllVoters(s.votersForSpawn, s.cfg, s.collector)
		return nil
	case enum.ManyVotersStrategy:
		s.voteStrategyForSpawn = votestrategy.NewManyVoters(s.votersForSpawn, s.cfg, s.collector)
		return nil
	case enum.AtLeastOneVoterStrategy:
		s.voteStrategyForSpawn = votestrategy.NewAtLeastOneVoter(s.votersForSpawn, s.cfg, s.collector)
		return nil
	default:
		return SpawnVoteStrategyWasNotFoundError
	}
}

func (s *Balancer) initVoteStrategyForClose() error {
	switch enum.VoteStrategy(s.cfg.VoteForCloseReqSenderStrategy) {
	case enum.AllVotersStrategy:
		s.voteStrategyForClose = votestrategy.NewAllVoters(s.votersForSpawn, s.cfg, s.collector)
		return nil
	case enum.ManyVotersStrategy:
		s.voteStrategyForClose = votestrategy.NewManyVoters(s.votersForSpawn, s.cfg, s.collector)
		return nil
	case enum.AtLeastOneVoterStrategy:
		s.voteStrategyForClose = votestrategy.NewAtLeastOneVoter(s.votersForSpawn, s.cfg, s.collector)
		return nil
	default:
		return CloseVoteStrategyWasNotFoundError
	}
}

func (s *Balancer) initVotersForSpawn() {
	s.votersForSpawn = []model.Voter{
		spawnvoter.ByRPS(s.cfg, s.collector),
		spawnvoter.ByInterval(s.cfg, s.collector),
		spawnvoter.ByMinWorkers(s.cfg, s.collector),
		spawnvoter.ByAvgDuration(s.cfg, s.collector),
	}
}

func (s *Balancer) initVotersForClose() {
	s.votersForSpawn = []model.Voter{
		closevoter.ByRPS(s.cfg, s.collector),
		closevoter.ByMaxWorkers(s.cfg, s.collector),
		closevoter.ByAvgDuration(s.cfg, s.collector),
	}
}

func (s *Balancer) IsMustBeSpawned() bool {
	return s.voteStrategyForSpawn.IsFor()
}

func (s *Balancer) IsMustBeClosed() bool {
	return s.voteStrategyForClose.IsFor()
}
