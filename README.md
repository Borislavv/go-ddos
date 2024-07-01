# DDOS/Stress tester
 
<img width="1705" alt="image" src="https://github.com/Borislavv/go-ddos/assets/50691459/03b4fd23-de52-491d-b17e-e7bea60ae6c1">

## Configuration:
```
type Config struct {
	URLs        []string `arg:"-u,env:urls,separate,required"` // example: -u http://localhost:8080 -u http://localhost:8081
	MaxRPS      int64    `arg:"env:MAX_RPS" default:"1000"`
	MinWorkers  int64    `arg:"env:MIN_WORKERS" default:"10"`
	MaxWorkers  int64    `arg:"env:MAX_WORKERS" default:"100"`
	MaxRequests int64    `arg:"env:MAX_REQUESTS"`

	// RequestHeaders add user request headers to each request (for example x-access-token).
	RequestHeaders map[string]string `arg:"-r,env:REQUEST_HEADERS"` // example: -r x-access-token=access-token x-country-code=us
	// AddTimestampToUrl is add unique timestamp in milliseconds value to each URL (commonly for avoid HTTP cache).
	AddTimestampToUrl bool `arg:"env:ADD_TIMESTAMP_TO_URL"`

	// Duration of application operation.
	Duration      string `arg:"env:DURATION" default:"1m"`
	DurationValue time.Duration

	// Stages is a number of parts by which will be separated output table.
	Stages int64 `arg:"-s,env:NUM_STAGES" default:"1"`
	// LogFile is a path to file into which will be redirected logs.
	LogFile string `arg:"-l,env:LOG_FILE"`
	// LogHeaders is a slice of headers which must be caught on request error.
	LogHeaders []string `arg:"-h,separate,env:CATCH_HEADERS"`
	// ExpectedResponseData is string which contains expected response data.
	// If it does not match, request will be marked as failed.
	ExpectedResponseData string `arg:"-e,env:EXPECTED_RESPONSE_DATA"`

	// PoolInitSize is httpclient pool init. size.
	PoolInitSize int64 `arg:"-i,env:HTTP_CLIENT_POOL_INIT_SIZE" default:"32"`
	// PoolMaxSize is httpclient pool max size.
	PoolMaxSize int64 `arg:"-m,env:HTTP_CLIENT_POOL_MAX_SIZE"   default:"10240"`

	// ToleranceCoefficient tell the application how much tolerance you can accept, since all values are very different
	// and achieving absolute results is very difficult (for example, due to network latency).
	// This will work as follows, for example, if we consider a strategy based on the average response time of
	// successful requests (spawn_by_avg_duration), then the calculation will be as follows:
	// average successful response time < target response time of successful requests * (1 - ToleranceCoefficient).
	// Example:
	// 	AvgSuccessResponseTime = 500ms
	//  TargetAvgSuccessRequestsDuration = 450ms
	//  AvgSuccessResponseTime * (1 - ToleranceCoefficient) = 450ms.
	// In this case, spawn will not continue.
	ToleranceCoefficient float64 `arg:"env:TOLERANCE_COEFFICIENT" default:"0.1"`

	// SpawnInterval is using for spawn_by_interval, spawn_by_avg_duration strategies.
	// It necessary for check the result of increasing/decreasing workers.
	SpawnInterval      string `arg:"-i,env:SPAWN_REQ_SENDER_INTERVAL" default:"1s"`
	SpawnIntervalValue time.Duration

	// TargetRPS used for spawn_by_rps strategy. If current RPS is under the target, spawn will be continued
	//	(used ToleranceCoefficient for determine aa space between from and to values).
	TargetRPS int64 `arg:"env:TARGET_RPS"`

	// TargetAvgSuccessRequestsDuration tells the target of average duration of success requests
	// which used into the spawn_by_avg_duration strategy.
	TargetAvgSuccessRequestsDuration      string        `arg:"env:TARGET_AVG_SUCCESS_REQUESTS_DURATION" default:"500ms"`
	TargetAvgSuccessRequestsDurationValue time.Duration // will contain the value after config parsing

	// SpawnVoteStrategy tells which approach will be used for spawn a new worker for send requests.
	// Options:
	// 	all 		 - means that all voters must vote for spawn a new worker.
	// 	many 		 - means that many voter must vote for spawn a new worker.
	// 	at_least_one - means that at least one voter must vote for spawn a new worker.
	SpawnVoteStrategy string `arg:"env:SPAWN_REQ_SENDER_STRATEGY" default:"at_least_one"`
	// CloseVoteStrategy tells which approach will be used for spawn a new worker for send requests.
	// Options:
	// 	all 		 - means that all voters must vote for close a worker.
	// 	many 		 - means that many voter must vote for close a worker.
	// 	at_least_one - means that at least one voter must vote for close a worker.
	CloseVoteStrategy string `arg:"env:CLOSE_REQ_SENDER_STRATEGY" default:"at_least_one"`

	// SpawnVoters tells which voters must be used for vote
	//	for spawn a new worker for send requests (allowed any combinations).
	//
	// Options:
	//	spawn_by_min_workers  	- for this case will be used MinWorkers configuration value.
	//		If number of active workers is under this value, them will spawn until it will not.
	//	spawn_by_rps			- for this case will be used current RPS (requests per second)
	//		for determine whether is necessary to spawn a new worker. For each 10 RPS will spawn a new worker.
	//	spawn_by_avg_duration 	- for this case will be use value of TargetAvgDuration (total reqs. duration / total reqs.).
	//		If the TargetAvgDuration is under the current average requests duration, the workers will be spawning.
	// 	spawn_by_interval	 	- for this case will be used SpawnInterval value.
	SpawnVoters []string `arg:"separate,env:REQ_SENDER_SPAWN_VOTERS"`
	// CloseVoters tells which voters must be used for vote
	//	for close a worker (allowed any combinations).
	//
	// Options:
	//	close_by_max_workers  - for this case will be used MinWorkers configuration value.
	//		If number of active workers is above this value, them will close one by one until it will not.
	//	close_by_rps			- for this case will be used current RPS (requests per second)
	//		for determine whether is necessary to close a worker.
	//		That is if the current RPS is above of (target RPS + 15%), the workers will be closing.
	//	close_by_avg_duration - for this case will be use value of TargetAvgDuration (total reqs. duration / total reqs.).
	//		If the TargetAvgDuration is above the current average requests duration, the workers will be closing.
	CloseVoters []string `arg:"separate,env:REQ_SENDER_CLOSE_VOTERS"`
}
```

## Example:
```
go run cmd/main.go --url="http://0.0.0.0:8080/api/v1/stressed-endpoint" \
 --maxrps=1000 \
 --duration=10m \
 --maxworkers=10 --minworkers=10 \
 --poolinitsize=32 --poolmaxsize=1024
```

## Disclaimer for DDOS Program

Please read this disclaimer carefully before using the DDOS program ("the software") developed by Glazunov Borislav.

The intention of the software: The DDOS program is developed solely for educational and research purposes. It is intended to understand and analyze the performance and resilience of networks and systems against distributed denial-of-service (DDOS) attacks.

Limitation of liability: The developer of the software, Glazunov Borislav, will not be liable for any damages, including but not limited to direct, indirect, special, incidental, or consequential damages or losses that occur due to the use or inability to use the software.

Use at your own risk: The user agrees to use the software at their own risk. The software is provided "as is" without warranty of any kind, either express or implied, including, but not limited to, the implied warranties of merchantability or fitness for a particular purpose.

Legal compliance: The user of the software is responsible for ensuring that their use of the software complies with all applicable laws and regulations. Unauthorized or illegal use of this software is strictly prohibited.

Modification of terms: Glazunov Borislav reserves the right to modify these terms at any time. Continued use of the software after such changes will constitute your consent to such changes.

By using the DDOS program, you acknowledge that you have read this disclaimer and agree to its terms.
