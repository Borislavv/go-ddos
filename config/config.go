package config

type Config struct {
	URL         string `arg:"-u,env:URL,required"`
	MaxRPS      int    `arg:"-r,env:MAX_RPS"         default:"10"`
	Percentiles int64  `arg:"-p,env:NUM_PERCENTILES" default:"4"`
	MaxWorkers  int64  `arg:"-w,env:MAX_WORKERS"     default:"5"`
	Duration    string `arg:"-d,env:DURATION"        default:"10m"`
	StdErrFile  string `arg:"-e,env:STDERR_FILE"`
	StdOutFile  string `arg:"-o,env:STDOUT_FILE"`
}
