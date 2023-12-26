package config

type Config struct {
	// URL is a string with target DDOS url.
	URL string `arg:"-u,env:URL,required"`
	// ExpectedResponseData is string which contains expected response data.
	// If it does not match, request will be marked as failed.
	ExpectedResponseData string `arg:"-e,env:EXPECTED_RESPONSE_DATA"`
	// MaxRPS is a maximum number of requests per one second.
	MaxRPS int `arg:"-r,env:MAX_RPS"                         default:"10"`
	// MaxWorkers is a maximum workers number which will be spawn.
	MaxWorkers int64 `arg:"-w,env:MAX_WORKERS"               default:"5"`
	// Duration is a string which contains duration of the DDOS execution.
	Duration string `arg:"-d,env:DURATION"                   default:"10m"`
	// Percentiles is a number of parts by which will be separated output table.
	Percentiles int64 `arg:"-p,env:NUM_PERCENTILES"          default:"1"`
	// LogFile is a path to file into which will be redirected logs.
	LogFile string `arg:"-l,env:LOG_FILE"                    default:"/dev/null"`
	// LogHeaders is a slice of headers which must be caught on request error.
	LogHeaders []string `arg:"-h,separate,env:CATCH_HEADERS"`
}
