// bootstrap
package coffeenet

import (
	"errors"
	"net"

	"github.com/coffeehc/logger"
)

//启动接口
type Bootstrap interface {
	//创建一个新的Server
	NewServer(netType, host string) *Server
	//创建一个信的Client
	NewClient(netType, host string) *Client
	//关闭多有的链接
	Close() error
	//获取统计接口信息
	GetStatInfo() StatInfo
}

type _bootstrap struct {
	//配置信息
	config *Config
	//设置连接参数
	connectionSetting func(conn net.Conn)
	//上下文工厂
	contextFactory *ContextFactory
	//统计信息
	handlerStat *HandlerStat
}

func (this *_bootstrap) GetStatInfo() StatInfo {
	return this
}

func (this *_bootstrap) GetHandlerStat() HandlerStat {
	return *this.handlerStat
}

func (this *_bootstrap) GetWorkRoutine() int {
	return len(this.contextFactory.workPool)
}

//初始化Bootstrap
func NewBootStrap(config *Config, contextFactory *ContextFactory, connectionSetting func(conn net.Conn)) Bootstrap {
	if config == nil {
		config = default_config
	}
	config.checkConfig()
	handlerStat := NewHandlerStat()
	handlerStat.StartHandlerStat()
	contextFactory.handlerStat = handlerStat
	contextFactory.orderHandler = config.OrderHandler
	contextFactory.workPool = make(chan int64, config.MaxConcurrentHandler)
	return &_bootstrap{config, connectionSetting, contextFactory, handlerStat}
}

//建立连接时处理连接参数并且创建上下文
func (this *_bootstrap) connection(conn net.Conn) (*Context, error) {
	//控制连接数
	if len(this.contextFactory.group) > this.config.MaxConnection {
		conn.Close()
		return nil, errors.New(logger.Warn("已经超出最大连接数,拒绝连接"))
	}
	logger.Info("成功创建连接%s->%s", conn.LocalAddr(), conn.RemoteAddr())
	//设置连接参数
	if this.connectionSetting != nil {
		this.connectionSetting(conn)
	}
	context := this.contextFactory.creatContext(conn)
	context.process()
	return context, nil
}

func (this *_bootstrap) Close() error {
	this.contextFactory.Close()
	return nil
}

func (this *_bootstrap) NewServer(netType, host string) *Server {
	return &Server{host, netType, this, nil}
}
func (this *_bootstrap) NewClient(netType, host string) *Client {
	return &Client{host, netType, nil, this}
}
