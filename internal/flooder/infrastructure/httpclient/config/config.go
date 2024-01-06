package httpclientconfig

type Config struct {
	PoolInitSize int64 `arg:"-i,env:HTTP_CLIENT_POOL_INIT_SIZE"   default:"32"`
	PoolMaxSize  int64 `arg:"-m,env:HTTP_CLIENT_POOL_MAX_SIZE"    default:"10240"`
}
