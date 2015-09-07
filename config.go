// config
package coffeenet

import "github.com/coffeehc/logger"

type Config struct {
	//最大连接数
	MaxConnection int
	//最大并发处理个数
	MaxConcurrentHandler int
	//是否顺序处理消息,默认false,即可以并发处理消息
	OrderHandler bool
}

var (
	default_config_maxConnecton         = 1000000
	default_config_maxConcurrentHandler = 1000000
	default_config                      = &Config{default_config_maxConnecton, default_config_maxConcurrentHandler, false}
)

//校验配置是否合法,并自动修复错误值
func (this *Config) checkConfig() {
	if this.MaxConnection <= 0 {
		logger.Warn("")
		this.MaxConnection = default_config_maxConnecton
	}
	if this.MaxConcurrentHandler <= 0 {
		this.MaxConcurrentHandler = default_config_maxConcurrentHandler
	}
}
