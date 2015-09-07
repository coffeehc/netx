// bootstrap
package coffeenet

import (
	"errors"
	"net"

	"github.com/coffeehc/logger"
)

//启动接口
type Bootstrap interface {
	//但连接创立的时候需要被调用
	NewServer(netType, host string) *Server
	NewClient(netType, host string) *Client
	Connection(conn net.Conn) (*Context, error)
	Close() error
}

type _bootStrap struct {
	//配置信息
	config *Config
	//设置连接参数
	connectionSetting func(conn net.Conn)
	//上下文工厂
	contextFactory *ContextFactory

	handlerStat *HanderStat
}

//初始化Bootstrap
func NewBootStrap(config *Config, contextFactory *ContextFactory, connectionSetting func(conn net.Conn)) Bootstrap {
	if config == nil {
		config = default_config
	}
	config.checkConfig()
	handlerStat := NewHanderStat()
	handlerStat.StartHanderStat()
	contextFactory.handlerStat = handlerStat
	contextFactory.orderHandler = config.OrderHandler
	contextFactory.workPool = make(chan int64, config.MaxConcurrentHandler)
	return &_bootStrap{config, connectionSetting, contextFactory, handlerStat}
}

//建立连接时处理连接参数并且创建上下文
func (this *_bootStrap) Connection(conn net.Conn) (*Context, error) {
	//控制连接数
	if len(this.contextFactory.group) > this.config.MaxConnection {
		conn.Close()
		return nil, errors.New(logger.Warn("已经超出最大连接数,拒绝连接"))
	}
	logger.Debug("成功创建连接%s->%s", conn.LocalAddr(), conn.RemoteAddr())
	//设置连接参数
	if this.connectionSetting != nil {
		this.connectionSetting(conn)
	}
	context := this.contextFactory.creatContext(conn)
	wait := make(chan bool)
	go context.process(wait)
	<-wait
	return context, nil
}
func (this *_bootStrap) Close() error {
	this.contextFactory.Close()
	return nil
}

func (this *_bootStrap) NewServer(netType, host string) *Server {
	return &Server{host, netType, this, nil}
}
func (this *_bootStrap) NewClient(netType, host string) *Client {
	return &Client{host, netType, nil, this}
}
