package ddos

type Config struct {
	RPC      int    `env:"RPC" envDefault:"100"`
	Workers  int    `env:"WORKERS" envDefault:"10"`
	Duration string `env:"DURATION" envDefault:"15m"`
}
