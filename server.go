// server
package coffeenet

import (
	"errors"
	"fmt"
	"net"

	"github.com/coffeehc/logger"
)

type Closer interface {
	Close() error
}

type Server struct {
	host      string
	netType   string
	bootstrap *_bootstrap
	closer    Closer
}

func (this *Server) GetBootstrap() Bootstrap {
	return this.bootstrap
}

func (this *Server) Bind() (err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = errors.New(logger.Error("bind出现错误:%s", errInfo))
			return
		}
	}()
	switch this.netType {
	case "tcp", "tcp4", "tcp6":
		this.serveTCP()
	default:
		panic("暂不支持TCP以外的协议")
	}
	return nil
}

func (this *Server) serveTCP() error {
	addr, err := net.ResolveTCPAddr(this.netType, this.host)
	if err != nil {
		logger.Error("转换TCP地址出现错误:%s", err)
		return err
	}
	listener, err := net.ListenTCP(this.netType, addr)
	if err != nil {
		return fmt.Errorf("bind出现错误:%s", err)
	}
	this.closer = listener
	logger.Info("已经bind:[%s]%s", this.netType, listener.Addr())
	go func(this *Server) {
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
				this.bootstrap.connection(conn)
			}
		}
	}(this)
	return nil
}

func (this *Server) Close() error {
	if this.closer != nil {
		return this.closer.Close()
	}
	return nil
}
