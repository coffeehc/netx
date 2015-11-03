// interface
package signal

import "github.com/coffeehc/coffeenet"

//信令处理接口
type SignalHandler interface {
	//处理信令
	Handle(context *coffeenet.Context, signal *Signal)
}

//信令路由
type RegeditSignal func(signalCode uint32, handler SignalHandler) error
