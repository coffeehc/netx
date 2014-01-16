// server
package coffeenet

import (
	"fmt"
	"net"

	"github.com/coffeehc/logger"
)

type Server struct {
	BootStrap
	listener net.Listener
	group    map[int]*ChannelHandlerContext
}

func NewServer(host string, netType string, workPoolSize int) *Server {
	server := new(Server)
	server.host = host
	server.netType = netType
	server.group = make(map[int]*ChannelHandlerContext)
	server.workConcurrent = workPoolSize
	server.initWorkPool()
	return server
}

func (this *Server) Bind() error {
	if this.open {
		return fmt.Errorf("Server已经启动")
	}
	leistener, err := net.Listen(this.netType, this.host)
	if err != nil {
		return fmt.Errorf("bind出现错误:%s", err)
	}
	this.open = true
	this.listener = leistener
	logger.Infof("已经bind:[%s]%s", this.netType, leistener.Addr())
	go func(this *Server) {
		listen := new(closeListen)
		listen.server = this
		for {
			conn, err := this.listener.Accept()
			if err != nil {
				logger.Warnf("Accept出现错误:%s", err)
			} else {
				channelHandlerContext := this.channelHandlerContextFactory.CreatChannelHandlerContext(conn, this.workPool)
				channelHandlerContext.AddListen(listen)
				//TODO 此处可以限制连接数
				go channelHandlerContext.handle()
			}
		}
	}(this)
	return nil
}

func (this *Server) Close() {
	for _, v := range this.group {
		v.Close()
	}
}

type closeListen struct {
	SimpleChannelListen
	server *Server
}

func (this *closeListen) OnClose(context *ChannelHandlerContext) {
	delete(this.server.group, context.id)
	context.listens = nil
}
