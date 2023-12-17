package ddos

type Config struct {
	RPS        int    `env:"RPS" envDefault:"300"`
	MaxWorkers int64  `env:"WORKERS" envDefault:"1000"`
	Duration   string `env:"DURATION" envDefault:"15m"`
}
