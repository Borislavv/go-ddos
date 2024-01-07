# ddos
 
![image](https://github.com/Borislavv/go-ddos/assets/50691459/4f752ffe-8dae-4aba-b753-3a154ffaa610)

## Configuration:
```
type Config struct {
	URL        string `arg:"env:URL,separate,required"`
	MaxRPS     int64  `arg:"env:MAX_RPS,required"`
	MinWorkers int64  `arg:"env:MIN_WORKERS,required"`
	MaxWorkers int64  `arg:"env:MAX_WORKERS,required"`

	// Duration of application operation.
	Duration      string `arg:"env:DURATION" default:"10m"`
	DurationValue time.Duration

	// Stages is a number of parts by which will be separated output table.
	Stages int64 `arg:"-s,env:NUM_STAGES"                    default:"5"`
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
}
```

Example:
```
go run cmd/main.go --url="https://seo-php-swoole.lux.kube.xbet.lan/api/v1/pagedata?group_id=285&ref_id=1&url=https%3A%2F%2Fjared.com%2Fes%2Flive%2Ffootball&geo=by&language=en&project[id]=285&domain=jared.com&stream=live&section=sport&sport[id]=10&timezone=1" \
 --maxrps=1000 \
 --duration=10m \
 --maxworkers=10 --minworkers=10 \
 --poolinitsize=32 --poolmaxsize=1024
```
