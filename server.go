package netx

import (
	"errors"
	"fmt"
	"net"

	"github.com/coffeehc/logger"
)

//Server net server
type Server interface {
	GetBootstrap() Bootstrap
	Bind() (err error)
	Close() error
}

//Server struct define
type _Server struct {
	host      string
	netType   string
	bootstrap *_bootstrap
	listener net.Listener
}

//GetBootstrap get base bootstrap
func (server *_Server) GetBootstrap() Bootstrap {
	return server.bootstrap
}

//Bind bind ip and port
func (server *_Server) Bind() (err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = errors.New(logger.Error("bind出现错误:%s", errInfo))
			return
		}
	}()
	switch server.netType {
	case "tcp", "tcp4", "tcp6":
		server.serveTCP()
	default:
		panic("暂不支持TCP以外的协议")
	}
	return nil
}

func (server *_Server) serveTCP() error {
	addr, err := net.ResolveTCPAddr(server.netType, server.host)
	if err != nil {
		logger.Error("转换TCP地址出现错误:%s", err)
		return err
	}
	listener, err := net.ListenTCP(server.netType, addr)
	if err != nil {
		return fmt.Errorf("bind出现错误:%s", err)
	}
	server.listener = listener
	logger.Info("已经bind:[%s]%s", server.netType, listener.Addr())
	go func(server *_Server) {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok {
					if opErr.Timeout() || opErr.Temporary() {
						continue
					} else {
						logger.Error("出现不可恢复的异常,关闭服务,%s", err)
						return
					}
				}
			} else {
				server.bootstrap.connection(conn)
			}
		}
	}(server)
	return nil
}

func (server *_Server) Close() error {
	if server.listener != nil {
		return server.listener.Close()
	}
	//TODO 需要考虑
	return nil
}
