package config

type Config struct {
	MaxRPS     int    `env:"MaxRPS" envDefault:"300"`
	MaxWorkers int64  `env:"WORKERS" envDefault:"1000"`
	Duration   string `env:"DURATION" envDefault:"15m"`
}
