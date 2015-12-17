// interface
package signal

import (
	"github.com/coffeehc/coffeenet"
	"net"
)

//信令处理接口
type SignalHandler interface {
	//处理信令
	Handle(context *coffeenet.Context, signal *Signal)
}

type SignalEngine interface {
	RegeditSignal(signalCode uint32, handler SignalHandler) error
	AddListen(name string,listen coffeenet.ContextListen)
	Connection(addr *net.TCPAddr) error
	Bind(addr *net.TCPAddr) (*coffeenet.Server, error)
	Close()
}

func NewSimpleSignal(signal uint32, data []byte) *Signal {
	return &Signal{Signal: &signal, Data: data}
}
