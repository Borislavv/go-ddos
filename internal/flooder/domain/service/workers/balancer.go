package workers

import (
	"context"
	"errors"
	"github.com/Borislavv/go-ddos/config"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/enum"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/vote"
	votestrategy "github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/vote/strategy"
	"github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter"
	closevoter "github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter/close"
	spawnvoter "github.com/Borislavv/go-ddos/internal/flooder/domain/service/workers/voter/spawn"
	logservice "github.com/Borislavv/go-ddos/internal/log/domain/service"
	statservice "github.com/Borislavv/go-ddos/internal/stat/domain/service"
)

var (
	UndefinedSpawnVoteStrategyWasGivenError = errors.New("undefined spawn vote strategy was given, check your input")
	UndefinedCloseVoteStrategyWasGivenError = errors.New("undefined close vote strategy was given, check your input")

	UndefinedSpawnVoterWasGivenError = errors.New("undefined spawn voter was given, check your input")
	UndefinedCloseVoterWasGivenError = errors.New("undefined spawn voter was given, check your input")
)

type BalancerService struct {
	ctx                  context.Context
	cfg                  *config.Config
	logger               logservice.Logger
	collector            statservice.Collector
	voteStrategyForSpawn vote.Strategy
	voteStrategyForClose vote.Strategy
	votersForSpawn       []voter.Voter
	votersForClose       []voter.Voter
}

func NewBalancerService(
	ctx context.Context,
	cfg *config.Config,
	logger logservice.Logger,
	collector statservice.Collector,
) *BalancerService {
	s := &BalancerService{
		ctx:       ctx,
		cfg:       cfg,
		logger:    logger,
		collector: collector,
	}

	if err := s.initVotersForSpawn(); err != nil {
		panic(err)
	}
	if err := s.initVotersForClose(); err != nil {
		panic(err)
	}

	if err := s.initVoteStrategyForSpawn(); err != nil {
		panic(err)
	}
	if err := s.initVoteStrategyForClose(); err != nil {
		panic(err)
	}

	return s
}

func (s *BalancerService) IsMustBeSpawned() bool {
	return s.voteStrategyForSpawn.IsFor()
}

func (s *BalancerService) IsMustBeClosed() bool {
	return s.voteStrategyForClose.IsFor()
}

func (s *BalancerService) initVoteStrategyForSpawn() error {
	switch enum.VoteStrategy(s.cfg.SpawnVoteStrategy) {
	case enum.AllVotersStrategy:
		s.voteStrategyForSpawn = votestrategy.NewAllVoters(s.votersForSpawn)

		s.logger.Println("workers.Balancer.initVoteStrategyForSpawn(): " +
			"'all' strategy was successfully sat up for spawn")

		return nil
	case enum.ManyVotersStrategy:
		s.voteStrategyForSpawn = votestrategy.NewManyVoters(s.votersForSpawn)

		s.logger.Println("workers.Balancer.initVoteStrategyForSpawn(): " +
			"'many' strategy was successfully sat up for spawn")

		return nil
	case enum.AtLeastOneVoterStrategy:
		s.voteStrategyForSpawn = votestrategy.NewAtLeastOneVoter(s.votersForSpawn)

		s.logger.Println("workers.Balancer.initVoteStrategyForSpawn(): " +
			"'at_least_one' strategy was successfully sat up for spawn")

		return nil
	default:
		s.logger.Printfln("workers.Balancer.initVoteStrategyForSpawn(): "+
			"error occurred due to undefined vote strategy '%v' was given", s.cfg.SpawnVoteStrategy)

		return UndefinedSpawnVoteStrategyWasGivenError
	}
}

func (s *BalancerService) initVoteStrategyForClose() error {
	switch enum.VoteStrategy(s.cfg.CloseVoteStrategy) {
	case enum.AllVotersStrategy:
		s.logger.Println("workers.Balancer.initVoteStrategyForClose(): " +
			"'all' strategy was successfully sat up for close")

		s.voteStrategyForClose = votestrategy.NewAllVoters(s.votersForClose)

		return nil
	case enum.ManyVotersStrategy:
		s.voteStrategyForClose = votestrategy.NewManyVoters(s.votersForClose)

		s.logger.Println("workers.Balancer.initVoteStrategyForClose(): " +
			"'many' strategy was successfully sat up for close")

		return nil
	case enum.AtLeastOneVoterStrategy:
		s.voteStrategyForClose = votestrategy.NewAtLeastOneVoter(s.votersForClose)

		s.logger.Println("workers.Balancer.initVoteStrategyForClose(): " +
			"'at_least_one' strategy was successfully sat up for close")

		return nil
	default:
		return UndefinedCloseVoteStrategyWasGivenError
	}
}

func (s *BalancerService) initVotersForSpawn() error {
	for _, spawnVoter := range s.cfg.SpawnVoters {
		switch enum.SpawnVoter(spawnVoter) {
		case enum.SpawnByMinWorkers:
			s.votersForSpawn = append(s.votersForSpawn, spawnvoter.NewByMinWorkers(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForSpawn(): " +
				"'spawn_by_min_workers' voter successfully sat up")
		case enum.SpawnByRPS:
			s.votersForSpawn = append(s.votersForSpawn, spawnvoter.NewByRPS(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForSpawn(): " +
				"'spawn_by_rps voter' successfully sat up")
		case enum.SpawnByAvgDuration:
			s.votersForSpawn = append(s.votersForSpawn, spawnvoter.NewByAvgDuration(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForSpawn(): " +
				"'spawn_by_avg_duration' voter successfully sat up")
		case enum.SpawnByInterval:
			s.votersForSpawn = append(s.votersForSpawn, spawnvoter.NewByInterval(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForSpawn(): " +
				"'spawn_by_interval' voter successfully sat up")
		default:
			s.logger.Printfln("workers.Balancer.initVotersForSpawn(): "+
				"undefined spawn voter '%v' was given", spawnVoter)

			return UndefinedSpawnVoterWasGivenError
		}
	}

	if len(s.votersForSpawn) == 0 {
		s.votersForSpawn = append(s.votersForSpawn, spawnvoter.NewByMinWorkers(s.cfg, s.collector))

		s.logger.Println("workers.Balancer.initVotersForSpawn(): " +
			"no one spawn voter was passed, will used default 'spawn_by_min_workers' voter")
	}

	return nil
}

func (s *BalancerService) initVotersForClose() error {
	for _, closeVoter := range s.cfg.CloseVoters {
		switch enum.CloseVoter(closeVoter) {
		case enum.CloseByMaxWorkers:
			s.votersForClose = append(s.votersForClose, closevoter.NewByMaxWorkers(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForClose(): " +
				"'close_by_max_workers' voter successfully sat up")
		case enum.CloseByAvgDuration:
			s.votersForClose = append(s.votersForClose, closevoter.NewByAvgDuration(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForClose(): " +
				"'close_by_avg_duration' voter successfully sat up")
		case enum.CloseByRPS:
			s.votersForClose = append(s.votersForClose, closevoter.NewByRPS(s.cfg, s.collector))

			s.logger.Println("workers.Balancer.initVotersForClose(): " +
				"'close_by_rps' voter successfully sat up")
		default:
			s.logger.Printfln("workers.Balancer.initVotersForClose(): "+
				"undefined close voter '%v' was given", closeVoter)

			return UndefinedCloseVoterWasGivenError
		}
	}

	if len(s.votersForClose) == 0 {
		s.logger.Println("workers.Balancer.initVotersForClose(): " +
			"no one close voter was passed, will used default 'close_by_max_workers' voter")

		s.votersForClose = append(s.votersForClose, closevoter.NewByMaxWorkers(s.cfg, s.collector))
	}

	return nil
}
