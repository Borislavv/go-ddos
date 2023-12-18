package config

type Config struct {
	MaxRPS     int    `env:"MAX_RPS" envDefault:"10"`
	MaxWorkers int64  `env:"WORKERS" envDefault:"1000"`
	Duration   string `env:"DURATION" envDefault:"15m"`
}
