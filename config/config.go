package config

type Config struct {
	URL         string `arg:"-u,env:URL,required"`
	MaxRPS      int    `arg:"-r,env:MAX_RPS"         default:"10"`
	Percentiles int64  `arg:"-p,env:NUM_PERCENTILES" default:"4"`
	MaxWorkers  int64  `arg:"-w,env:MAX_WORKERS"     default:"5"`
	Duration    string `arg:"-d,env:DURATION"        default:"10m"`
	LogFile     string `arg:"-l,env:LOG_FILE"        default:"/dev/null"`
}
