package config

type Config struct {
	MaxRPS      int    `arg:"-r,env:MAX_RPS"         default:"10"`
	Percentiles int    `arg:"-p,env:NUM_PERCENTILES" default:"4"`
	MaxWorkers  int64  `arg:"-w,env:MAX_WORKERS"     default:"5"`
	Duration    string `arg:"-d,env:DURATION"        default:"10m"`
	URL         string `arg:"-u,env:URL,required"`
}
