package signal

import (
	"net"

	"github.com/coffeehc/netx"
)

//Handler 信令处理接口
type Handler interface {
	//处理信令
	Handle(context netx.ConnContext, signal *Signal)
}

//Engine 信令处理引擎
type Engine interface {
	RegisterSignal(signalCode uint32, handler Handler) error
	AddListen(name string, listen netx.ContextListen)
	Connection(addr *net.TCPAddr) error
	Bind(addr *net.TCPAddr) (netx.Server, error)
	Close()
	GetBootStrap() netx.Bootstrap
}

//NewSimpleSignal 创建一个低级别的 Signal
func NewSimpleSignal(signal uint32, data []byte) *Signal {
	return &Signal{Signal: &signal, Data: data}
}
