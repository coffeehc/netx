// bootstrap
package signal

import (
	"fmt"
	"net"
	"sync"

	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/coffeenet/protocol"
	"github.com/golang/protobuf/proto"
)

func NewSignalBootstrap(netSetting func(conn net.Conn)) (coffeenet.Bootstrap, RegeditSignal) {
	factroy := &initFactory{make(map[uint32]SignalHandler, 0), new(sync.Mutex)}
	contextFactory := coffeenet.NewContextFactory(factroy.initContextFactory)
	return coffeenet.NewBootStrap(nil, contextFactory, netSetting), factroy.regeditHandler

}

type initFactory struct {
	signals map[uint32]SignalHandler
	mutex   *sync.Mutex
}

func (this *initFactory) initContextFactory(context *coffeenet.Context) {
	protocols := []coffeenet.Protocol{protocol.NewLengthFieldProtocol(4)}
	protocols = append(protocols, protocol.NewProtoBufProcotol(func() proto.Message { return new(Signal) }))
	context.SetProtocols(protocols)
	context.SetHandler(&netHandler{this})
}

//注册Handler
func (this *initFactory) regeditHandler(signalCode uint32, handler SignalHandler) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.signals[signalCode]; ok {
		return fmt.Errorf("code[%x]对应的处理接口已经存在.")
	}
	this.signals[signalCode] = handler
	return nil
}

//获取Handler
func (this *initFactory) getHandler(signalCode uint32) SignalHandler {
	return this.signals[signalCode]
}
