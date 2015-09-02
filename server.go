// server
package coffeenet

import (
	"fmt"
	"net"

	"github.com/coffeehc/logger"
)

type ServerFD interface {
	Close() error
}

type Server struct {
	BootStrap
	host                         string
	netType                      string
	fd                           ServerFD
	channelHandlerContextFactory *ChannelHandlerContextFactory
}

func (this *Server) GetHanderStat() HanderStat {
	return this.channelHandlerContextFactory.GetHanderStat()
}

func (this *Server) GetWorkRuntine() int {
	return len(this.workPool)
}

func NewServer(netType string, host string, workPoolSize int) *Server {
	server := new(Server)
	server.host = host
	server.netType = netType
	server.group = make(map[int32]*ChannelHandlerContext)
	server.workConcurrent = workPoolSize
	server.init()
	return server
}

func (this *Server) Bind(setting func(conn net.Conn)) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("bind出现错误:%s", err)
			return
		}
	}()
	this.open.Do(func() {
		logger.Info("bind%s", this.host)
		switch this.netType {
		case "tcp", "tcp4", "tcp6":
			this.serveTCP(setting)
		default:
			panic("暂不支持TCP以外的协议")
		}
	})
	return nil
}

func (this *Server) serveTCP(setting func(conn net.Conn)) error {
	addr, err := net.ResolveTCPAddr(this.netType, this.host)
	if err != nil {
		logger.Error("转换TCP地址出现错误:%s", err)
		return err
	}
	listener, err := net.ListenTCP(this.netType, addr)
	if err != nil {
		return fmt.Errorf("bind出现错误:%s", err)
	}
	this.fd = listener
	logger.Info("已经bind:[%s]%s", this.netType, listener.Addr())
	go func(this *Server) {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && (opErr.Timeout() || opErr.Temporary()) {
					continue
				}
				logger.Warn("Accept出现错误:%s", err)
			} else {
				if setting != nil {
					setting(conn)
				}
				logger.Debug("已经建立连接:%s->%s", conn.LocalAddr(), conn.RemoteAddr())
				channelHandlerContext := this.channelHandlerContextFactory.CreatChannelHandlerContext(conn, this.workPool)
				//TODO 此处可以限制连接数
				go channelHandlerContext.handle()
			}
		}
	}(this)
	return nil
}

func (this *Server) SetChannelHandlerContextFactory(factory *ChannelHandlerContextFactory) {
	this.channelHandlerContextFactory = factory
	this.channelHandlerContextFactory.bootStrap = &this.BootStrap
}
