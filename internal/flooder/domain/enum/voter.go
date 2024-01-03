package enum

type SpawnVoter string

const (
	SpawnByRPS         SpawnVoter = "spawn_by_rps"
	SpawnByInterval    SpawnVoter = "spawn_by_interval"
	SpawnByMinWorkers  SpawnVoter = "spawn_by_min_workers"
	SpawnByAvgDuration SpawnVoter = "spawn_by_avg_duration"
)

func (s SpawnVoter) String() string {
	return string(s)
}

type CloseVoter string

const (
	CloseByRPS         CloseVoter = "close_by_rps"
	CloseByMaxWorkers  CloseVoter = "close_by_max_workers"
	CloseByAvgDuration CloseVoter = "close_by_avg_duration"
)

func (s CloseVoter) String() string {
	return string(s)
}
