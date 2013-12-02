// server
package coffeenet

import (
	"fmt"
	"logger"
	"net"
)

type Server struct {
	BootStrap
	listener net.Listener
}

func NewServer(host string, netType string) *Server {
	server := new(Server)
	server.host = host
	server.netType = netType
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
		for {
			conn, err := this.listener.Accept()
			if err != nil {
				logger.Warnf("Accept出现错误:%s", err)
			} else {
				channelHandlerContext := this.channelHandlerContextFactory.CreatChannelHandlerContext(conn)
				go channelHandlerContext.handle()
			}
		}
	}(this)
	return nil
}
