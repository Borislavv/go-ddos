package httpclientconfig

type Config struct {
	PoolInitSize int `arg:"-c,env:HTTP_CLIENT_POOL_INIT_SIZE"   default:"32"`
	PoolMaxSize  int `arg:"-c,env:HTTP_CLIENT_POOL_MAX_SIZE"    default:"10240"`
}
