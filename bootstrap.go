package netx

import (
	"errors"
	"net"

	"github.com/coffeehc/logger"
)

//Bootstrap 启动接口
type Bootstrap interface {
	//NewServer 创建一个新的Server
	NewServer(netType, host string) Server
	//NewClient 创建一个信的Client
	NewClient(netType, host string) Client
	//Close 关闭多有的链接
	Close() error
}

type _bootstrap struct {
	//配置信息
	config *Config
	//设置连接参数
	connectionSetting func(conn net.Conn)
	//上下文工厂
	contextFactory *_ContextFactory
	//统计信息
}

func (bs *_bootstrap) GetWorkRoutine() int {
	return len(bs.contextFactory.workPool)
}

//NewBootStrap 初始化Bootstrap
func NewBootStrap(config *Config, initContextFunc InitConnContextFunc, connectionSetting func(conn net.Conn)) Bootstrap {
	if config == nil {
		config = defaultConfig
	}
	config.checkConfig()
	contextFactory := newContextFactory(initContextFunc)
	contextFactory.syncHandler = config.SyncHandler
	contextFactory.workPool = make(chan int64, config.MaxConcurrentHandler)
	return &_bootstrap{
		config:            config,
		connectionSetting: connectionSetting,
		contextFactory:    contextFactory,
	}
}

//建立连接时处理连接参数并且创建上下文
func (bs *_bootstrap) connection(conn net.Conn) (ConnContext, error) {
	//控制连接数
	if len(bs.contextFactory.group) > bs.config.MaxConnection {
		conn.Close()
		return nil, errors.New(logger.Warn("已经超出最大连接数,拒绝连接"))
	}
	logger.Info("成功创建连接%s->%s", conn.LocalAddr(), conn.RemoteAddr())
	//设置连接参数
	if bs.connectionSetting != nil {
		bs.connectionSetting(conn)
	}
	connContext := bs.contextFactory.createConnContext(conn)
	//TODO 考虑超时
	connContext.Start(bs.contextFactory.rootCxt)
	return connContext, nil
}

func (bs *_bootstrap) Close() error {
	bs.contextFactory.close()
	return nil
}

func (bs *_bootstrap) NewServer(netType, host string) Server {
	return &_Server{
		host:host,
		netType:netType,
		bootstrap:bs,
	}
}
func (bs *_bootstrap) NewClient(netType, host string) Client {
	return &_Client{
		host:host,
		netType:netType,
		bootstrap:bs,
	}
}
