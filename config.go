package netx

//Config net config
type Config struct {
	//最大连接数
	MaxConnection int `json:"max_connection"`
	//最大并发处理个数
	MaxConcurrentHandler int `json:"max_concurrent_handler"`
	//是否顺序处理消息,默认false,即可以并发处理消息
	SyncHandler bool `json:"sync_handler"`
}

var (
	defaultConfigMaxConnecton         = 1000000
	defaultConfigMaxConcurrentHandler = 1000000
	defaultConfig                     = &Config{defaultConfigMaxConnecton, defaultConfigMaxConcurrentHandler, false}
)

//校验配置是否合法,并自动修复错误值
func (config *Config) checkConfig() {
	if config.MaxConnection <= 0 {
		config.MaxConnection = defaultConfigMaxConnecton
	}
	if config.MaxConcurrentHandler <= 0 {
		config.MaxConcurrentHandler = defaultConfigMaxConcurrentHandler
	}
}
