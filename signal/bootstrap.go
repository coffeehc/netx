// bootstrap
package signal

import (
	"fmt"
	"net"
	"sync"

	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/coffeenet/protocol"
	"github.com/coffeehc/logger"
	"github.com/golang/protobuf/proto"
	"time"
)

func NewSignalBootstrap(config *coffeenet.Config, netSetting func(conn net.Conn), compressProtocol coffeenet.Protocol, listens map[string]coffeenet.ContextListen) SignalEngine {
	factroy := new(initFactory)
	bootstrap := coffeenet.NewBootStrap(config, coffeenet.NewContextFactory(factroy.initContextFactory), netSetting)
	factroy.signals = make(map[uint32]SignalHandler, 0)
	factroy.mutex = new(sync.Mutex)
	factroy.bootstrap = bootstrap
	factroy.listens = listens
	if factroy.listens == nil {
		factroy.listens = make(map[string]coffeenet.ContextListen)
	}
	return factroy

}

type initFactory struct {
	compressProtocol coffeenet.Protocol
	signals          map[uint32]SignalHandler
	mutex            *sync.Mutex
	bootstrap        coffeenet.Bootstrap
	listens          map[string]coffeenet.ContextListen
}

func (this *initFactory) GetBootStrap() coffeenet.Bootstrap {
	return this.bootstrap
}

func (this *initFactory) AddListen(name string, listen coffeenet.ContextListen) {
	if _, ok := this.listens[name]; ok {
		logger.Error("listen[%s]已经注册,不能再次注册", name)
		return
	}
	this.listens[name] = listen
}

func (this *initFactory) initContextFactory(context *coffeenet.Context) {
	for name, listen := range this.listens {
		context.AddListen(name, listen)
	}
	protocols := []coffeenet.Protocol{protocol.NewLengthFieldProtocol(4)}
	if this.compressProtocol != nil {
		protocols = append(protocols, this.compressProtocol)
	}
	protocols = append(protocols, protocol.NewProtoBufProcotol(func() proto.Message { return new(Signal) }))
	context.SetProtocols(protocols)
	context.SetHandler(&netHandler{factory: this, remortAddr: context.RemortAddr()})
}

//注册Handler
func (this *initFactory) RegisterSignal(signalCode uint32, handler SignalHandler) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.signals[signalCode]; ok {
		return fmt.Errorf("code[0x%X]对应的处理接口已经存在.", signalCode)
	}
	this.signals[signalCode] = handler
	logger.Debug("注册 Code[0x%X]=>%#T", signalCode, handler)
	return nil
}

//获取Handler
func (this *initFactory) getHandler(signalCode uint32) SignalHandler {
	return this.signals[signalCode]
}

func (this *initFactory) Connection(addr *net.TCPAddr) error {
	client := this.bootstrap.NewClient(addr.Network(), addr.String())
	return client.Connect(5 * time.Second)
}
func (this *initFactory) Bind(addr *net.TCPAddr) (*coffeenet.Server, error) {
	server := this.bootstrap.NewServer(addr.Network(), addr.String())
	if err := server.Bind(); err != nil {
		return nil, err
	}
	return server, nil
}

func (this *initFactory) Close() {
	this.bootstrap.Close()
}
