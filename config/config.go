package config

type Config struct {
	// URL is a string with target DDOS url.
	URL string `arg:"-u,separate,env:URL,required"`
	// ExpectedResponseData is string which contains expected response data.
	// If it does not match, request will be marked as failed.
	ExpectedResponseData string `arg:"-e,env:EXPECTED_RESPONSE_DATA"`
	// MaxRPS is a maximum number of requests per one second.
	MaxRPS int `arg:"-r,env:MAX_RPS"                         default:"10"`
	// InitWorkersNum is an initial number of workers which will be spawn
	// (if you didn't specify any other spawn strategies, then it will be the max workers number).
	InitWorkersNum int64 `arg:"-w,env:MAX_WORKERS"           default:"5"`
	// Duration is a string which contains duration of the DDOS execution.
	Duration string `arg:"-d,env:DURATION"                   default:"10m"`
	// Stages is a number of parts by which will be separated output table.
	Stages int64 `arg:"-s,env:NUM_STAGES"                    default:"1"`
	// LogFile is a path to file into which will be redirected logs.
	LogFile string `arg:"-l,env:LOG_FILE"`
	// LogHeaders is a slice of headers which must be caught on request error.
	LogHeaders []string `arg:"-h,separate,env:CATCH_HEADERS"`

	// PoolInitSize is httpclient pool init. size.
	PoolInitSize int `arg:"-i,env:HTTP_CLIENT_POOL_INIT_SIZE" default:"32"`
	// PoolMaxSize is httpclient pool max size.
	PoolMaxSize int `arg:"-m,env:HTTP_CLIENT_POOL_MAX_SIZE"   default:"10240"`

	// ReqSenderSpawnInterval is using for spawn_by_interval strategy.
	ReqSenderSpawnInterval string `arg:"-i,env:SPAWN_REQ_SENDER_INTERVAL"     default:"1m"`
	// VoteForSpawnReqSenderStrategy tells which approach will be used for spawn a new worker for send requests.
	// Options:
	// 	all 		 - means that all voters must vote for spawn a new worker.
	// 	many 		 - means that many voter must vote for spawn a new worker.
	// 	at_least_one - means that at least one voter must vote for spawn a new worker.
	VoteForSpawnReqSenderStrategy string `arg:"env:SPAWN_REQ_SENDER_STRATEGY" default:"at_least_one"`
	// VoteForCloseReqSenderStrategy tells which approach will be used for spawn a new worker for send requests.
	// Options:
	// 	all 		 - means that all voters must vote for close a worker.
	// 	many 		 - means that many voter must vote for close a worker.
	// 	at_least_one - means that at least one voter must vote for close a worker.
	VoteForCloseReqSenderStrategy string `arg:"env:CLOSE_REQ_SENDER_STRATEGY" default:"at_least_one"`
	// ReqSenderSpawnVoters tells which voters must be used for vote
	//	for spawn a new worker for send requests (allowed any combinations).
	//
	// Options:
	//	spawn_by_max_workers  	- for this case will be used InitWorkersNum configuration value.
	//		If number of active workers is under this value, them will spawn until it will not.
	//	spawn_by_rps			- for this case will be used current RPS (requests per second)
	//		for determine whether is necessary to spawn a new worker. For each 10 RPS will spawn a new worker.
	//	spawn_by_avg_duration 	- for this case will be use value of TargetAvgDuration (total reqs. duration / total reqs.).
	//		If the TargetAvgDuration is under the current average requests duration, the workers will be spawning.
	// 	spawn_by_interval	 	- for this case will be used ReqSenderSpawnInterval value.
	ReqSenderSpawnVoters []string `arg:"separate,env:REQ_SENDER_SPAWN_VOTERS"`
	// ReqSenderCloseVoters tells which voters must be used for vote
	//	for close a worker (allowed any combinations).
	//
	// Options:
	//	close_by_max_workers  - for this case will be used InitWorkersNum configuration value.
	//		If number of active workers is above this value, them will close one by one until it will not.
	//	close_by_rps			- for this case will be used current RPS (requests per second)
	//		for determine whether is necessary to close a worker.
	//		That is if the current RPS is above of (target RPS + 15%), the workers will be closing.
	//	close_by_avg_duration - for this case will be use value of TargetAvgDuration (total reqs. duration / total reqs.).
	//		If the TargetAvgDuration is above the current average requests duration, the workers will be closing.
	ReqSenderCloseVoters []string `arg:"separate,env:REQ_SENDER_CLOSE_VOTERS"`
}
