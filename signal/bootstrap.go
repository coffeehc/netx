package signal

import (
	"fmt"
	"net"
	"sync"

	"time"

	"context"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
	"github.com/coffeehc/netx/protocol"
	"github.com/golang/protobuf/proto"
)

//NewSignalBootstrap create a signal bootstrap
func NewSignalBootstrap(config *netx.Config, netSetting func(conn net.Conn), compressProtocol netx.Protocol, listens map[string]netx.ContextListen) Engine {
	factroy := new(initFactory)
	bootstrap := netx.NewBootStrap(config, factroy.initContextFactory, netSetting)
	factroy.signals = make(map[uint32]Handler, 0)
	factroy.mutex = new(sync.Mutex)
	factroy.bootstrap = bootstrap
	factroy.listens = listens
	if factroy.listens == nil {
		factroy.listens = make(map[string]netx.ContextListen)
	}
	factroy.compressProtocol = compressProtocol
	return factroy

}

type initFactory struct {
	compressProtocol netx.Protocol
	signals          map[uint32]Handler
	mutex            *sync.Mutex
	bootstrap        netx.Bootstrap
	listens          map[string]netx.ContextListen
}

func (f *initFactory) GetBootStrap() netx.Bootstrap {
	return f.bootstrap
}

func (f *initFactory) AddListen(name string, listen netx.ContextListen) {
	if _, ok := f.listens[name]; ok {
		logger.Error("listen[%s]已经注册,不能再次注册", name)
		return
	}
	f.listens[name] = listen
}

func (f *initFactory) initContextFactory(cxt context.Context, context netx.ConnContext) {
	for name, listen := range f.listens {
		context.AddListen(name, listen)
	}
	protocols := []netx.Protocol{protocol.NewLengthFieldProtocol(4)}
	if f.compressProtocol != nil {
		protocols = append(protocols, f.compressProtocol)
	}
	protocols = append(protocols, protocol.NewProtoBufProcotol(func() proto.Message { return new(Signal) }))
	context.SetProtocols(protocols...)
	context.SetHandler(&netHandler{factory: f, remortAddr: context.RemoteAddr()})
}

//注册Handler
func (f *initFactory) RegisterSignal(signalCode uint32, handler Handler) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.signals[signalCode]; ok {
		return fmt.Errorf("code[0x%X]对应的处理接口已经存在", signalCode)
	}
	f.signals[signalCode] = handler
	logger.Debug("注册 Code[0x%X]=>%#T", signalCode, handler)
	return nil
}

//获取Handler
func (f *initFactory) getHandler(signalCode uint32) Handler {
	return f.signals[signalCode]
}

func (f *initFactory) Connection(addr *net.TCPAddr) error {
	client := f.bootstrap.NewClient(addr.Network(), addr.String())
	return client.Connect(5 * time.Second)
}
func (f *initFactory) Bind(addr *net.TCPAddr) (netx.Server, error) {
	server := f.bootstrap.NewServer(addr.Network(), addr.String())
	if err := server.Bind(); err != nil {
		return nil, err
	}
	return server, nil
}

func (f *initFactory) Close() {
	f.bootstrap.Close()
}
